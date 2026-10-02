# Giobby su Coolify — piano di implementazione

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** produzione di giobby-backend su Coolify (solo VPN), deploy solo con CI verde, migrazioni automatiche, ripartenza dopo reboot.

**Architecture:** flusso `lavoro → dev → main`; la CI su `main` porta il branch `release` in fast-forward e Coolify costruisce da `release` con `docker-compose.prod.yml`. Un servizio `migrate` (stessa immagine del backend, `cmd/migrate`) applica le migrazioni versionate prima che parta il backend. Postgres è quello condiviso di cloud4job-infra.

**Tech Stack:** Go 1.22 (pgx v4 già in `go.mod`), PostgreSQL 17, Docker Compose, GitHub Actions, Coolify, nginx.

**Spec:** `docs/superpowers/specs/2026-10-02-coolify-deploy-design.md`

## Global Constraints

- Nessuna attribuzione AI nei commit (niente `Co-Authored-By`, niente "Generated with…"); autore `Francesco Halla <francesco.halla@cloud4job.com>`.
- Conferma esplicita dell'utente prima di ogni commit, push, PR, modifica a GitHub o comando che scrive sul server.
- Mai push diretti o `--force` su `dev`, `main`, `release`; mai `--no-verify`.
- Nessun segreto nel repo, nei log o nei messaggi: in locale solo valori sintetici.
- Porta pubblicata solo `10.8.0.1:2001:80` (api-explorer); mai `0.0.0.0`.
- Immagini base con versione fissa (niente `latest`); Postgres 17 ovunque.
- Go non è installato sul Mac: i comandi Go si eseguono con
  `docker run --rm -v "$PWD/backend":/src -w /src golang:1.22 <cmd>` (alias `GO` sotto).
- Nuova dipendenza Go: nessuna (si usa pgx v4 già presente).

## Review Focus

1. Password con caratteri speciali (`@`, `/`, `:`) nel DSN: `PostgreSQLConnectionString` non fa escaping → le password generate sono esadecimali (`openssl rand -hex 24`) e `cmd/migrate` costruisce il DSN con `net/url`. Test in Task 2.
2. Due deploy ravvicinati che lanciano `migrate` insieme: l'advisory lock serializza, una sola applica. Test in Task 2.
3. Migrazione già applicata e poi modificata nel repo, o `migrate` che fallisce per altri motivi: `migrate` esce 1 e il backend nuovo non parte. Ci si aspetta che il backend già in esecuzione resti su: è da verificare, non è garantito. Test in Task 2 + smoke Task 4 Step 5.
4. Variabile mancante in Coolify: il deploy fallisce subito con il nome della variabile, invece di partire con i default di sviluppo (`dev-jwt-secret-change-me`, `postgres/postgres`). Test in Task 4.
5. Merge su `main` con CI rossa: `release` non si muove. Verifica in Task 5 (job `release` con `needs`) e Task 8.

---

### Task 1: Regole del repository e segreti fuori da git

**Files:**
- Create: `.githooks/pre-push`, `.github/CONTRIBUTING.md`, `backend/.env.example`
- Modify: `.gitignore`
- Untrack: `backend/.env` (il file locale resta)

- [ ] **Step 1:** `.githooks/pre-push` (eseguibile): come quello di try2catch-me-backend, ma blocca `refs/heads/main`, `refs/heads/dev`, `refs/heads/release` con il messaggio `Push diretto su <branch> bloccato: pusha un branch e apri una pull request (vedi .github/CONTRIBUTING.md).`
- [ ] **Step 2:** verifica: `git config core.hooksPath .githooks`, poi `printf 'refs/heads/x 0 refs/heads/main 0\n' | .githooks/pre-push; echo $?` → `1`; con `refs/heads/feat/x` → `0`; ripetere per `dev` e `release` → `1`.
- [ ] **Step 3:** `.github/CONTRIBUTING.md`: flusso `feat|fix|chore|docs/…` → PR → `dev` → PR → `main`; `release` gestito solo dal bot; `git config core.hooksPath .githooks` una volta per clone; CI verde + review prima del merge; nessuna attribuzione AI; nessun segreto; link al CONTRIBUTING di try2catch-me-docs come origine delle regole.
- [ ] **Step 4:** `backend/.env.example` con tutte le chiavi di `backend/.env` più `MIGRATION_DB_USER=giobby`, `MIGRATION_DB_PASSWORD=dev-migration-password`, `MIGRATION_APP_ROLE=giobby_app`; `POSTGRES_DB_USER=giobby_app`, `POSTGRES_DB_PASSWORD=dev-app-password`, `POSTGRES_DB_NAME=giobby`, `JWT_SECRET=dev-only-jwt-secret`. Aggiornare il `backend/.env` locale con gli stessi valori (resta non tracciato).
- [ ] **Step 5:** `.gitignore`: aggiungere `backend/.env` e `.env`; `git rm --cached backend/.env`.
- [ ] **Step 6:** verifica: `git ls-files backend/.env` → vuoto; `git check-ignore backend/.env` → stampa il percorso.
- [ ] **Step 7:** commit (dopo conferma): `Move secrets out of git and add contribution rules`.

### Task 2: Runner delle migrazioni

**Files:**
- Move: `db/init_scripts/*.sql` → `backend/internal/db/migrations/` (`git mv`, contenuto invariato)
- Create: `backend/internal/db/migrate.go`, `backend/internal/db/migrate_test.go`, `backend/cmd/migrate/main.go`
- Modify: `backend/Dockerfile`

**Interfaces:**
- Produces: `func Migrate(ctx context.Context, conn *pgx.Conn, appRole string) (applied int, err error)` in package `db`; errori esportati `ErrChecksumMismatch`, `ErrUnknownVersion`; binario `/root/migrate` nell'immagine backend; output `Database schema is current.` (exit 0).
- Env di `cmd/migrate`: `POSTGRES_DB_HOST`, `POSTGRES_DB_PORT` (default `5432`), `POSTGRES_DB_NAME`, `POSTGRES_DB_SSLMODE` (default `disable`), `MIGRATION_DB_USER`, `MIGRATION_DB_PASSWORD`, `MIGRATION_APP_ROLE` (vuoto = nessun grant).

- [ ] **Step 1:** `git mv db/init_scripts/*.sql backend/internal/db/migrations/`; `db/init_scripts/README.md` → `backend/internal/db/migrations/README.md` aggiornato: "versione = prefisso numerico; mai modificare un file già applicato; nuova modifica = nuovo file con prefisso maggiore".
- [ ] **Step 2: test che falliscono** in `migrate_test.go` (saltati se `GIOBBY_TEST_DATABASE_URL` è vuota; ogni test ricrea lo schema `public` e i ruoli di test):
  - `TestMigrateAppliesAllThenNothing`: prima chiamata `applied == numero di file .sql embedded`, `schema_migrations` ha quel numero di righe; seconda chiamata `applied == 0`.
  - `TestMigrateRejectsChangedChecksum`: dopo la prima, `UPDATE schema_migrations SET checksum='x' WHERE version=10`; `Migrate` → `errors.Is(err, ErrChecksumMismatch)`; nessuna riga cambiata.
  - `TestMigrateRejectsUnknownVersion`: riga con `version=99999` → `ErrUnknownVersion`.
  - `TestMigrateConcurrent`: due `Migrate` in goroutine su connessioni diverse → somma degli `applied` = numero di file, nessun errore.
  - `TestAppRoleCanWriteDataNotSchema`: con `appRole="giobby_test_app"`, connesso come quel ruolo: `INSERT` su `users` riesce; `CREATE TABLE t(x int)` fallisce; `INSERT INTO schema_migrations` fallisce.
  - `TestMigrationDSNEscapesPassword` (unit, senza DB) in `cmd/migrate`: password `a@b/c:d` → DSN che `pgx.ParseConfig` legge con la stessa password.
- [ ] **Step 3:** eseguire e vedere fallire:
  ```bash
  docker network create giobby-test
  docker run -d --rm --name giobby-test-pg --network giobby-test -e POSTGRES_PASSWORD=test postgres:17
  docker run --rm --network giobby-test -v "$PWD/backend":/src -w /src \
    -e GIOBBY_TEST_DATABASE_URL=postgres://postgres:test@giobby-test-pg:5432/postgres?sslmode=disable \
    golang:1.22 go test ./internal/db/ ./cmd/migrate/
  ```
  Atteso: errore di compilazione (`Migrate` non definita).
- [ ] **Step 4:** implementare `Migrate`: `//go:embed migrations/*.sql`; file ordinati per versione (prefisso numerico); una transazione; `SELECT pg_advisory_xact_lock(<costante>)`; `CREATE TABLE IF NOT EXISTS public.schema_migrations(version int PRIMARY KEY, name text NOT NULL, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`; verifica checksum SHA-256 delle versioni registrate e versioni sconosciute prima di applicare; applica le mancanti in ordine; grant al ruolo app (spec §3) con `pgx.Identifier{appRole}.Sanitize()`; `REVOKE ALL ON schema_migrations FROM <appRole>` + `GRANT SELECT`; commit.
- [ ] **Step 5:** `cmd/migrate/main.go`: legge l'env, DSN con `net/url` (`url.UserPassword`), `pgx.Connect` con timeout 30 s, chiama `Migrate`, stampa `Applied N migration(s).` e `Database schema is current.`; su errore stampa l'errore (mai il DSN) ed esce 1.
- [ ] **Step 6:** rieseguire il comando dello Step 3 → tutti i test PASS; poi `GO go vet ./...` e `GO go test ./...` → PASS (test esistenti compresi).
- [ ] **Step 7:** `backend/Dockerfile`: `golang:1.22.12-alpine` e `alpine:3.20` (versioni fisse), compila anche `go build -o migrate ./cmd/migrate/main.go`, copia `/app/migrate` in `/root/`. Verifica: `docker build -f backend/Dockerfile -t giobby-backend:test . && docker run --rm --entrypoint ls giobby-backend:test /root` → contiene `main migrate swagger.json`.
- [ ] **Step 8:** pulizia: `docker rm -f giobby-test-pg; docker network rm giobby-test`.
- [ ] **Step 9:** commit (dopo conferma): `Run versioned database migrations from a dedicated command`.

### Task 3: Compose di sviluppo locale

**Files:**
- Modify: `docker-compose.yml`
- Create: `dev/postgres-init/roles.sql`
- Delete: `db/Dockerfile`, `db/README.md` (cartella `db/` vuota dopo Task 2)

**Interfaces:** consuma `/root/migrate` (Task 2) e le chiavi di `backend/.env.example` (Task 1).

- [ ] **Step 1:** `dev/postgres-init/roles.sql`: `CREATE ROLE giobby LOGIN PASSWORD 'dev-migration-password'`, `CREATE ROLE giobby_app LOGIN PASSWORD 'dev-app-password'`, `CREATE DATABASE giobby OWNER giobby`, `REVOKE CONNECT ON DATABASE giobby FROM PUBLIC`, `GRANT CONNECT … TO giobby_app`. Commento in testa: solo sviluppo, valori sintetici.
- [ ] **Step 2:** `docker-compose.yml`: `db` → `image: postgres:17.<patch>-alpine` (tag di patch fisso: il più recente che `docker pull` trova, stesso nel job CI), monta `./dev/postgres-init:/docker-entrypoint-initdb.d:ro`, `POSTGRES_PASSWORD` sintetica, porta `127.0.0.1:5432:5432`; nuovo servizio `migrate` (build backend, `command: ["./migrate"]`, `env_file: ./backend/.env`, `POSTGRES_DB_HOST: db`, `depends_on: db: service_healthy`, `restart: "no"`); `backend` con `depends_on: migrate: service_completed_successfully`; tolte tutte le righe `image: alexgiura/...`; `api-explorer` invariato tranne `depends_on: backend: service_healthy` e healthcheck backend `wget -qO- http://127.0.0.1:8080/healthz`.
- [ ] **Step 3:** `git rm -r db/`.
- [ ] **Step 4: smoke** (avvisare l'utente prima: cancella il database locale):
  ```bash
  docker compose down -v && docker compose up -d --build
  docker compose logs migrate | grep "Database schema is current."
  curl -fsS localhost:8080/healthz          # OK
  curl -s -o /dev/null -w "%{http_code}\n" localhost:8080/api/users   # 401
  docker compose exec db psql -U giobby -d giobby -Atc "select count(*) from schema_migrations"   # 35
  docker compose up -d --force-recreate migrate && docker compose logs migrate | grep "Applied 0 migration"
  ```
  L'endpoint protetto dello smoke va scelto da `swagger.json` (uno qualsiasi sotto `/api/` che richieda auth).
- [ ] **Step 5:** commit (dopo conferma): `Run migrations in local compose on PostgreSQL 17`.

### Task 4: Compose di produzione

**Files:**
- Create: `docker-compose.prod.yml`, `scripts/check-prod-compose.sh`

- [ ] **Step 1: test** `scripts/check-prod-compose.sh`:
  - con tutte le variabili fittizie esportate, `docker compose -f docker-compose.prod.yml config -q` → exit 0;
  - per ognuna di `POSTGRES_DB_HOST POSTGRES_DB_NAME POSTGRES_DB_USER POSTGRES_DB_PASSWORD JWT_SECRET MIGRATION_DB_USER MIGRATION_DB_PASSWORD MIGRATION_APP_ROLE`: senza quella variabile → exit ≠ 0 e stderr contiene il nome;
  - `config --format json`: nessun servizio `db`, nessun volume, l'unica porta pubblicata è `host_ip 10.8.0.1`, `published 2001`, `target 80`;
  - stampa `prod compose OK`.
- [ ] **Step 2:** eseguire → fallisce (file assente).
- [ ] **Step 3:** scrivere `docker-compose.prod.yml` come da spec §4 (`migrate`, `backend`, `api-explorer`; obbligatorie con `${VAR:?missing VAR}`; logging `local` 20m × 5; `restart: unless-stopped` su backend e api-explorer, `"no"` su migrate). Commento in testa come quello di consuntiva: solo Coolify, branch `release`, solo VPN, Postgres condiviso.
- [ ] **Step 4:** eseguire lo script → `prod compose OK`.
- [ ] **Step 5: smoke reale** contro il Postgres locale del Task 3: copia temporanea in scratchpad con `10.8.0.1` sostituito da `127.0.0.1`, rete `giobby-backend_app` come external, variabili come in `.env.example`, `POSTGRES_DB_HOST=db`; `up -d --build` → `curl -fsS 127.0.0.1:2001/healthz` → `OK`; `docker inspect` dei container: health `healthy`; poi rilanciare `up -d` con `MIGRATION_DB_PASSWORD` sbagliata → `up` esce ≠ 0, `migrate` exit 1 e il backend avviato prima è ancora `healthy` e risponde (se invece viene fermato, annotarlo nella guida come limite); poi `down` (senza `-v` sul compose locale).
- [ ] **Step 6:** commit (dopo conferma): `Add Coolify production compose (VPN only)`.

### Task 5: CI, rilascio e guardia

**Files:**
- Delete: `.github/workflows/giobby.yml`
- Create: `.github/workflows/ci.yml`, `.github/workflows/branch-guard.yml`

- [ ] **Step 1:** `ci.yml` come da spec §2. Job `go`: `actions/setup-go@v5` con `go-version-file: backend/go.mod`, servizio Postgres con lo stesso tag del Task 3 (password sintetica), `GIOBBY_TEST_DATABASE_URL` puntato al servizio; passi `test -z "$(gofmt -l .)"`, `go vet ./...`, `go build ./...`, `go test ./...`. Job `docker`: build `backend/Dockerfile` e `api-explorer`, `scripts/check-prod-compose.sh`. Job `release`: `if: github.event_name == 'push' && github.ref == 'refs/heads/main'`, `needs: [go, docker]`, `permissions: contents: write`, `concurrency: release` (no cancel), checkout con `fetch-depth: 0`, `git push origin "$GITHUB_SHA:refs/heads/release"` (senza `--force`).
- [ ] **Step 2:** `branch-guard.yml`: su `push` a `dev`/`main` la logica di "Guardia main" di try2catch (commit da PR unita, 3 tentativi); su `push` a `release` fallisce sempre con `::error::release si aggiorna solo dal bot dopo CI verde su main`.
- [ ] **Step 3:** se `gofmt -l` in locale (`GO gofmt -l .`) segnala file esistenti, fermarsi e chiedere all'utente se formattarli in un commit separato o togliere il controllo: non decidere da soli.
- [ ] **Step 4:** verifica: `docker run --rm -v "$PWD":/repo -w /repo rhysd/actionlint:1.7.7` → nessun errore.
- [ ] **Step 5:** commit (dopo conferma): `Replace SSH deploy with CI and green-only release branch`.

### Task 6: Guida operativa

**Files:**
- Create: `docs/DEPLOY-COOLIFY.md`

- [ ] **Step 1:** guida sul modello di quella di consuntiva: requisiti; creazione ruoli e DB (spec §6.3, password `openssl rand -hex 24`); sysctl e dnsmasq (§5, §6.4); risorsa Coolify (branch `release`, file, rete predefinita, tabella variabili con nome/valore/obbligatoria); verifica (§6.6); aggiornamenti futuri e migrazioni; diagnosi (checksum mismatch, migrate fallito, porta `10.8.0.1`, `release` fermo = CI rossa su `main`); attività fuori scope.
- [ ] **Step 2:** verifica: ogni variabile `${…}` di `docker-compose.prod.yml` compare nella tabella (`grep -o '\${[A-Z_]*' docker-compose.prod.yml | sort -u` confrontato con la guida).
- [ ] **Step 3:** commit (dopo conferma): `Document the Coolify production setup`.

### Task 7: Pubblicazione su GitHub

Ogni passo richiede conferma esplicita.

- [ ] **Step 1:** creare `dev` da `main` sul remoto: `gh api repos/Cloud4Job/giobby-backend/git/refs -f ref=refs/heads/dev -f sha=$(git rev-parse origin/main)`.
- [ ] **Step 2:** `git push -u origin chore/coolify-deploy`; PR verso `dev` (cosa cambia, perché, verifiche, punti da guardare, nessuna attribuzione AI).
- [ ] **Step 3:** CI della PR verde (`gh pr checks`); review dell'altro sviluppatore; merge dall'utente.

### Task 8: Server, Coolify e messa in produzione

Ogni comando che scrive richiede conferma esplicita. Serve la VPN attiva.

- [ ] **Step 1 (sola lettura):** `ssh root@10.8.0.1`: nome e versione del container Postgres condiviso; `giobby`/`giobby_app` assenti; `ss -tlnp | grep ':2001'` vuoto; `sysctl net.ipv4.ip_nonlocal_bind`.
- [ ] **Step 2:** ruoli e database come da guida; verifica `\l giobby` e `\du giobby*`. Le password vanno all'utente fuori dal repo e dalla chat (file temporaneo sul server con `umask 077`, letto dall'utente e poi cancellato).
- [ ] **Step 3:** prova sysctl: `docker run --rm -d -p 10.8.0.254:9999:80 --name bindtest nginx:1.27-alpine` fallisce; `sysctl -w net.ipv4.ip_nonlocal_bind=1`; stesso comando parte; `docker rm -f bindtest`; persistenza in `/etc/sysctl.d/99-nonlocal-bind.conf`.
- [ ] **Step 4:** riga dnsmasq, `systemctl restart dnsmasq`, `ss -tulpn | grep dnsmasq` solo su `10.8.0.1`.
- [ ] **Step 5:** PR a cloud4job-infra: sysctl nella sezione riavvio, riga giobby nella tabella "Solo VPN".
- [ ] **Step 6:** PR `dev` → `main`, CI verde, merge dall'utente; verificare che `release` esista e punti al commit di `main`.
- [ ] **Step 7 (utente):** risorsa Coolify secondo la guida, Deploy.
- [ ] **Step 8:** verifiche della spec §6.6, poi redeploy senza modifiche → `Applied 0 migration(s).`; restart dello stack da Coolify → riparte.
