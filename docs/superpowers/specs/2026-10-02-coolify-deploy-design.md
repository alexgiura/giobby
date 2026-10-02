# Giobby su Coolify (produzione, solo VPN) — design

Data: 2026-10-02

## Obiettivo

Portare la produzione di giobby-backend su Coolify (server Hetzner di
[cloud4job-infra](https://github.com/Cloud4Job/cloud4job-infra)), sostituendo
il deploy attuale via SSH + Docker Hub, con queste garanzie:

- una CI rossa non arriva mai in produzione;
- le modifiche allo schema del database si applicano da sole a ogni deploy;
- dopo un riavvio del server lo stack riparte da solo.

Regole di lavoro: quelle dello starter kit try2catch
([CONTRIBUTING](https://github.com/Cloud4Job/try2catch-me-docs/blob/main/CONTRIBUTING.md),
`ai/regole-comuni.md`), adattate a giobby.

## Decisioni prese

| Tema | Decisione |
|---|---|
| Flusso branch | `feat/…`/`fix/…`/`chore/…` → PR → `dev` → PR → `main` (= produzione) |
| Deploy | Coolify costruisce da branch tecnico `release`, che solo il bot porta al commit di `main` dopo CI verde |
| Deploy attuale (SSH + Docker Hub) | Rimosso dalla action |
| Database | Postgres condiviso di cloud4job-infra (17); database `giobby`, ruoli `giobby` (owner/migrazioni) e `giobby_app` (API, solo dati) |
| Schema | Migrazioni versionate con checksum, eseguite da un servizio `migrate` prima del backend a ogni deploy |
| Dati iniziali | Database pulito, nessuna migrazione dal vecchio server |
| Esposizione | Solo VPN: `api-explorer` su `10.8.0.1:2001`, nome `giobby.cloud4job.com` via dnsmasq |
| Riavvio server | `net.ipv4.ip_nonlocal_bind=1` sul server; `restart: unless-stopped` |
| Segreti | Solo variabili Coolify; `backend/.env` esce dal repo; JWT e password nuovi |
| Commit | Nessuna attribuzione AI |

`dev` per ora non ha un ambiente Coolify (solo integrazione).

## 1. Flusso e garanzia "CI rossa ≠ produzione"

L'organizzazione è su GitHub Free con repo privato: la branch protection non è
disponibile (API: "Upgrade to GitHub Pro…"). La garanzia quindi non può stare
nel pulsante Merge, sta nel deploy:

```
lavoro ──PR──▶ dev ──PR──▶ main ──CI verde──▶ release ──webhook──▶ Coolify
```

- Push su `main` → CI completa → solo se tutti i job sono verdi, il job
  `release` esegue `git push origin <sha>:refs/heads/release` (fast-forward,
  mai `--force`) con `GITHUB_TOKEN` (`contents: write`). Se il fast-forward
  non è possibile il job fallisce e la produzione resta ferma.
- Coolify è collegato al branch `release`; il push del bot genera il webhook
  GitHub verso la porta 8000 (già aperta agli IP `hooks`). Il firewall non cambia.
- Push da `GITHUB_TOKEN` non avviano altri workflow: ogni esecuzione di un
  workflow su `release` è quindi un push manuale.
- `.githooks/pre-push` blocca i push diretti su `dev`, `main`, `release`.
- Workflow **Guardia branch** (da "Guardia main" di try2catch): su `dev`/`main`
  fallisce se il commit non arriva da una PR unita; su `release` fallisce
  sempre (solo il bot può aggiornarlo, e il bot non lo attiva).
- `.github/CONTRIBUTING.md`: flusso, hook, regola sull'attribuzione AI.

## 2. CI (`.github/workflows/ci.yml`, sostituisce `giobby.yml`)

Trigger: `pull_request`, `push` su `dev` e `main`, `workflow_dispatch`.
`permissions: contents: read` (il solo job `release` ha `contents: write`).

- `go`: Go da `backend/go.mod`; `gofmt -l` vuoto, `go vet ./...`,
  `go build ./...`, `go test ./...` in `backend/` con un servizio `postgres:17`
  per i test di integrazione delle migrazioni.
- `docker`: build delle immagini (`backend` con `cmd/migrate`, `api-explorer`)
  e `docker compose -f docker-compose.prod.yml config` con variabili fittizie.
- `release`: solo su push in `main`, `needs: [go, docker]`, fast-forward di `release`.

## 3. Migrazioni

Modello: `internal/database/migrate.go` di try2catch-me-backend, esteso a più versioni.

- `db/init_scripts/NNN_nome.sql` → `backend/internal/db/migrations/NNN_nome.sql`
  (`git mv`, contenuto invariato). Versione = prefisso numerico. Incluse nel
  binario con `embed`. Unica fonte dello schema, locale e produzione.
- `backend/internal/db/migrate.go`: `Migrate(ctx, conn, appRole string) error`
  - una transazione, `pg_advisory_xact_lock` sulla stessa connessione;
  - tabella `public.schema_migrations(version int PK, name text, checksum text, applied_at timestamptz)`;
  - applica in ordine solo le versioni mancanti e le registra con SHA-256;
  - errore se una versione applicata ha checksum diverso o non esiste più nei file;
  - se `appRole` non è vuoto: `GRANT USAGE` sullo schema, `SELECT, INSERT, UPDATE, DELETE`
    su tutte le tabelle, `USAGE, SELECT` sulle sequenze, a ogni esecuzione (idempotente).
    Nessun privilegio DDL né accesso a `schema_migrations` in scrittura.
- `backend/cmd/migrate/main.go`: usa `POSTGRES_DB_HOST/PORT/NAME/SSLMODE`
  più `MIGRATION_DB_USER`, `MIGRATION_DB_PASSWORD`, `MIGRATION_APP_ROLE`.
  Stampa `Database schema is current.` ed esce 0; altrimenti esce 1.
- Il backend **non** migra all'avvio; resta come oggi.
- Test di integrazione (Postgres reale, saltati se manca `GIOBBY_TEST_DATABASE_URL`):
  prima esecuzione applica tutto; seconda non cambia nulla; checksum alterato →
  errore senza modifiche; esecuzioni concorrenti → una sola applica; il ruolo
  app può fare DML ma non `CREATE TABLE`.
- `backend/Dockerfile`: compila anche `./cmd/migrate`; immagini base con
  versione fissa (no `alpine:latest`).

## 4. Compose

### `docker-compose.prod.yml` (nuovo, solo Coolify)

- `migrate`: build `backend/Dockerfile`, `command: ["./migrate"]`,
  `restart: "no"`. Variabili obbligatorie (`${VAR:?}`): `POSTGRES_DB_HOST`,
  `POSTGRES_DB_NAME`, `MIGRATION_DB_USER`, `MIGRATION_DB_PASSWORD`,
  `MIGRATION_APP_ROLE`.
- `backend`: stessa build, `depends_on: migrate: service_completed_successfully`.
  Obbligatorie: `POSTGRES_DB_HOST`, `POSTGRES_DB_NAME`, `POSTGRES_DB_USER`,
  `POSTGRES_DB_PASSWORD`, `JWT_SECRET`. Fisse: `SERVER_PORT=8080`,
  `POSTGRES_DB_PORT=5432`, `POSTGRES_DB_SSLMODE=disable`,
  `ENVIRONMENT=production`, `SWAGGER_PATH=/root/swagger.json`. Opzionali con
  default: `APP_TIMEZONE`, `ACCESS_TOKEN_TTL_SECONDS`, `REFRESH_TOKEN_TTL_DAYS`,
  `CORS_ALLOWED_ORIGINS`, `DEBUG_MODE=false`. Nessuna porta. Healthcheck
  `wget -qO- http://127.0.0.1:8080/healthz`.
- `api-explorer`: build `./api-explorer`, `depends_on: backend: service_healthy`,
  porta `"10.8.0.1:2001:80"` (mai `0.0.0.0`).
- Tutti i servizi a lunga vita: `restart: unless-stopped`, logging `local` 20m × 5.
- Nessun Postgres, nessun volume.

### `docker-compose.yml` (sviluppo locale)

- `db`: `postgres:17` (come produzione) con `dev/postgres-init/roles.sql` che
  crea `giobby` (owner) e `giobby_app` con password sintetiche di sviluppo.
- `migrate` e `backend` come in produzione, valori da `backend/.env`.
- Tolte le righe `image: alexgiura/...:prod`. Rimossa la cartella `db/`.
- Chi ha già il volume locale `giobby_pgdata` (Postgres 16) deve ricrearlo
  (`docker compose down -v`): documentato.

### Segreti

- `git rm --cached backend/.env`, in `.gitignore`; `backend/.env.example` con
  valori sintetici. Il vecchio `JWT_SECRET` resta nella storia git: in
  produzione si usano valori nuovi.

## 5. Riavvio del server

Dopo un reboot Docker può partire prima di WireGuard: `10.8.0.1` non esiste,
il bind della porta fallisce e il container resta fermo (successo a consuntiva
il 2026-09-21).

- `/etc/sysctl.d/99-nonlocal-bind.conf` con `net.ipv4.ip_nonlocal_bind=1`:
  il bind su `10.8.0.1` riesce anche prima che l'interfaccia esista. Vale anche
  per consuntiva.
- Prova prima di applicarlo: container usa e getta legato a un IP inesistente
  (es. `10.8.0.254`) fallisce senza sysctl e parte con sysctl; poi rimosso.
- Documentato con una PR al README di cloud4job-infra: sezione "Attenzione dopo
  un riavvio del server" e riga di giobby nella tabella "Solo VPN".

## 6. Messa in produzione

Ogni comando che scrive su GitHub o sul server richiede conferma esplicita.

1. **Repo**: modifiche sopra sul branch `chore/coolify-deploy`; verifiche locali
   (gofmt, vet, build, test con Postgres 17, compose locale funzionante,
   `docker compose -f docker-compose.prod.yml config`).
2. **GitHub**: creare `dev` da `main`; PR `chore/coolify-deploy` → `dev`, CI
   verde, review; PR `dev` → `main`; la CI su `main` crea `release`.
3. **Server** (VPN, `docker exec` sul Postgres condiviso): controlli in sola
   lettura (container, versione, assenza di `giobby`/`giobby_app`, porta 2001
   libera); `CREATE USER giobby …`, `CREATE DATABASE giobby OWNER giobby`,
   `CREATE USER giobby_app …`, `GRANT CONNECT`, `REVOKE CONNECT … FROM PUBLIC`.
   Lo schema lo crea il servizio `migrate` al primo deploy.
4. **Server**: sysctl (§5) e riga dnsmasq `address=/giobby.cloud4job.com/10.8.0.1`.
5. **Coolify** (l'utente, dalla guida): progetto `giobby` → risorsa Docker
   Compose `Cloud4Job/giobby-backend`, branch `release`,
   `docker-compose.prod.yml`, "Connect To Predefined Network", variabili,
   auto-deploy; Deploy.
6. **Verifica** (VPN): log `migrate` con `Database schema is current.`;
   `curl http://giobby.cloud4job.com:2001/healthz` → `OK`; API Explorer
   apribile; `/api/...` protetto → 401; secondo deploy senza modifiche → `migrate`
   non applica nulla; riavvio dello stack da Coolify → riparte.

## `docs/DEPLOY-COOLIFY.md`

Guida operativa con i passi 3–6, aggiornamenti futuri (nuova migrazione =
nuovo file `NNN_…sql`, mai modificare uno già applicato), diagnosi
(`checksum mismatch`, `migrate` fallito → backend non parte, porta `10.8.0.1`).

## Rischi e limiti accettati

- Un merge su `main` con CI rossa è possibile (GitHub Free) ma non va in
  produzione: `release` resta indietro finché `main` non torna verde.
- Le migrazioni non hanno rollback automatico: si scrive una nuova migrazione.
- Il riavvio reale del server non si prova in questo lavoro: la prova è quella
  del §5 sul bind.

## Fuori scope

- Spegnimento del vecchio server SSH e rimozione dei segreti GitHub non più
  usati (`DOCKERHUB_*`, `SSH_*`, `PRODUCTION_SERVER_IP`).
- Ambiente Coolify per `dev`.
