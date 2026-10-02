# Giobby su Coolify (cloud4job-infra)

Produzione di giobby-backend come risorsa Docker Compose di Coolify, con
`docker-compose.prod.yml`, sul server Hetzner di
[cloud4job-infra](https://github.com/Cloud4Job/cloud4job-infra). Database nel **Postgres condiviso**
(container `s5zzoofp0c63av31bhi1n1lo`, database `giobby`). Raggiungibile **solo in VPN, in HTTP**, su
`http://giobby.cloud4job.com:2001` (porta legata a `10.8.0.1`, nessun Traefik, nessun TLS).

Design: [`docs/superpowers/specs/2026-10-02-coolify-deploy-design.md`](superpowers/specs/2026-10-02-coolify-deploy-design.md).

## 1. Come arriva una modifica in produzione

```
feat/… ──PR──▶ dev ──PR──▶ main ──CI verde──▶ release ──webhook──▶ Coolify
```

- Coolify fa il deploy dal branch **`release`**, che sposta solo la CI (job `release` in
  `.github/workflows/ci.yml`) dopo una CI tutta verde su `main`. Con la CI rossa `release` resta
  fermo e la produzione non cambia.
- A ogni deploy Coolify ricostruisce le immagini e avvia, in ordine:
  1. **`migrate`**: applica le migrazioni pendenti (`backend/internal/db/migrations/`) come `giobby`
     e riassegna i permessi a `giobby_app`; poi esce;
  2. **`giobby-backend`**: parte solo se `migrate` è uscito con 0; usa `giobby_app` (solo dati, niente DDL).
     Il nome è volutamente specifico: sulla rete condivisa di Coolify un nome generico come `backend`
     potrebbe risolvere al container di un altro stack;
  3. **`api-explorer`**: parte quando il backend è `healthy`; espone UI e proxy `/api/`, `/health`,
     `/healthz` su `10.8.0.1:2001`.

## 2. Requisiti

| Cosa | Requisito |
|---|---|
| PostgreSQL | 18.6 (Postgres condiviso di cloud4job-infra; CI e compose locale usano la stessa versione) |
| Ruolo `giobby` | LOGIN, proprietario del database `giobby`, non superuser: esegue le migrazioni |
| Ruolo `giobby_app` | LOGIN, non superuser, non proprietario di nulla: i permessi li dà `migrate` |
| Password | esadecimali (`openssl rand -hex 24`): il backend non fa escaping nel DSN |
| Server | `net.ipv4.ip_nonlocal_bind=1` (§4), riga dnsmasq per `giobby.cloud4job.com` |

## 3. Database (una volta)

Sul server (`ssh root@10.8.0.1`, VPN attiva), `PG=s5zzoofp0c63av31bhi1n1lo`.

1. **Controlli in sola lettura**:
   ```bash
   docker exec $PG psql -U postgres -Atc "select current_setting('server_version_num')::int/10000"   # 17
   docker exec $PG psql -U postgres -Atc "select rolname from pg_roles where rolname like 'giobby%'"  # vuoto
   docker exec $PG psql -U postgres -Atc "select datname from pg_database where datname='giobby'"      # vuoto
   ss -tlnp | grep ':2001 '                                                                             # vuoto
   ```
2. **Password** in un file leggibile solo da root, da copiare poi nelle variabili Coolify:
   ```bash
   umask 077
   printf 'MIGRATION_DB_PASSWORD=%s\nPOSTGRES_DB_PASSWORD=%s\nJWT_SECRET=%s\n' \
     "$(openssl rand -hex 24)" "$(openssl rand -hex 24)" "$(openssl rand -hex 32)" > /root/giobby-secrets.env
   ```
3. **Ruoli e database** (le password lette dal file, mai scritte a mano nella shell):
   ```bash
   . /root/giobby-secrets.env
   docker exec -i $PG psql -U postgres -v ON_ERROR_STOP=1 \
     -v owner_pw="$MIGRATION_DB_PASSWORD" -v app_pw="$POSTGRES_DB_PASSWORD" <<'SQL'
   CREATE ROLE giobby LOGIN PASSWORD :'owner_pw';
   CREATE ROLE giobby_app LOGIN PASSWORD :'app_pw';
   CREATE DATABASE giobby OWNER giobby;
   REVOKE CONNECT ON DATABASE giobby FROM PUBLIC;
   GRANT CONNECT ON DATABASE giobby TO giobby_app;
   SQL
   ```
   Lo schema **non** si crea a mano: lo crea `migrate` al primo deploy.
4. Dopo aver copiato i valori in Coolify (§5): `rm /root/giobby-secrets.env`.

## 4. Server: riavvio e DNS interno (una volta)

**Ripartenza dopo un riavvio.** Se Docker parte prima di WireGuard, `10.8.0.1` non esiste ancora e il
bind della porta fallisce (successo a consuntiva il 21/9). Con `ip_nonlocal_bind` il bind riesce lo stesso:

```bash
# prova: senza sysctl fallisce, con sysctl parte
docker run --rm -d -p 10.8.0.254:9999:80 --name bindtest nginx:1.27-alpine; docker rm -f bindtest
sysctl -w net.ipv4.ip_nonlocal_bind=1
docker run --rm -d -p 10.8.0.254:9999:80 --name bindtest nginx:1.27-alpine && docker rm -f bindtest
# persistenza
echo 'net.ipv4.ip_nonlocal_bind=1' > /etc/sysctl.d/99-nonlocal-bind.conf
```

**DNS interno.** In `/etc/dnsmasq.d/internal.conf`:

```
address=/giobby.cloud4job.com/10.8.0.1
```

poi `systemctl restart dnsmasq` e `ss -tulpn | grep dnsmasq` (deve ascoltare solo su `10.8.0.1`).

## 5. Risorsa Coolify (una volta)

Dal pannello (`http://coolify.cloud4job.com:8000`, in VPN):

1. Progetto **giobby** → *New Resource* → **Git Repository (with GitHub App)** → App `cloud4job-coolify`
   (installata sull'organizzazione con accesso a tutti i repository) → `Cloud4Job/giobby-backend`.
2. **Branch `release`** (non `main`), **Build pack `Docker Compose`** (il default è Railpack/Nixpacks),
   Docker Compose location **`/docker-compose.prod.yml`** (il default `/docker-compose.yaml` è il compose locale),
   Base directory `/`, Watch paths vuoto.
3. Impostazioni: **Connect To Predefined Network** attivo; **nessun dominio** su nessun servizio (Coolify può
   generarne uno `…sslip.io` pubblico: va cancellato); auto-deploy attivo.
4. Pagina delle variabili, **Normal View**:
   - **Build secrets → "Standard build arguments"** (vedi §5.1);
   - per **ogni** variabile, dall'ingranaggio: **Build time → "Not available during build"**, Runtime disponibile.

   | Variabile | Valore | Obbligatoria |
   |---|---|---|
   | `POSTGRES_DB_HOST` | `s5zzoofp0c63av31bhi1n1lo` | sì |
   | `POSTGRES_DB_NAME` | `giobby` | sì |
   | `POSTGRES_DB_USER` | `giobby_app` | sì |
   | `POSTGRES_DB_PASSWORD` | da `/root/giobby-secrets.env` | sì |
   | `JWT_SECRET` | da `/root/giobby-secrets.env` (almeno 32 caratteri, il backend lo controlla) | sì |
   | `MIGRATION_DB_USER` | `giobby` | sì |
   | `MIGRATION_DB_PASSWORD` | da `/root/giobby-secrets.env` | sì |
   | `MIGRATION_APP_ROLE` | `giobby_app` | sì |
   | `APP_TIMEZONE` | default `Europe/Bucharest` | no |
   | `ACCESS_TOKEN_TTL_SECONDS` | default `3600` | no |
   | `REFRESH_TOKEN_TTL_DAYS` | default `7` | no |
   | `CORS_ALLOWED_ORIGINS` | default `http://giobby.cloud4job.com:2001` | no |
   | `DEBUG_MODE` | default `false` | no |

   Se una variabile obbligatoria è vuota il deploy si ferma con `required variable <NOME> is missing a value`.
5. **Prima del deploy** controllare dal database di Coolify che nessuna variabile sia build-time (sola lettura):
   ```bash
   docker exec coolify-db psql -U coolify -d coolify -Atc "select key, is_buildtime from environment_variables
     where resourceable_type like '%Application' and not is_preview and resourceable_id =
     (select id from applications where uuid = '<uuid della risorsa>') order by 1"   # tutte 'f'
   ```
6. **Deploy** (il primo deploy richiede che `release` esista: lo crea la CI al primo merge su `main`).
   Nel log: al massimo `Added 2 ARG declarations` (solo `COOLIFY_FQDN`). Se compaiono di più, una variabile è
   ancora build-time: i suoi valori finiscono nella cronologia dell'immagine (`docker history`) e vanno ruotati.
7. Cancellare `/root/giobby-secrets.env` dopo aver copiato i valori.

### 5.1 Insidie di Coolify (v4.3.23, verificate il 2026-10-02)

| Comportamento | Effetto | Come evitarlo |
|---|---|---|
| Con **Build secrets = "Docker BuildKit secrets"** e più servizi con `build:`, Coolify riscrive ogni Dockerfile riusando un buffer che accoda (`dockerfile_content`, `append` di default in `ExecuteRemoteCommand`) | il Dockerfile di un servizio riceve quello dei precedenti: build di `api-explorer` con il Dockerfile del backend (`"/swagger.json": not found`) | Build secrets = **"Standard build arguments"** |
| Le variabili **Build time** diventano `ARG` iniettati nei Dockerfile | con "Standard build arguments" i valori restano in chiaro nella cronologia delle immagini | **tutte** le variabili "Not available during build" (nessuna serve al build) |
| La **vista developer** (testo `CHIAVE=valore`) ricrea le variabili con Build time attivo | un segreto aggiornato da lì torna build-time | aggiornare i segreti **dall'ingranaggio** nella Normal View; ricontrollare con il §5 punto 5 |
| `${VAR:?messaggio}` nel compose | Coolify salva il messaggio come **valore** (`JWT_SECRET=missing JWT_SECRET`) | usare `${VAR:?}` senza messaggio (lo verifica `scripts/check-prod-compose.sh`); il backend rifiuta comunque in produzione un `JWT_SECRET` segnaposto o sotto i 32 caratteri |
| Le variabili citate nel compose non si possono cancellare dall'interfaccia | "Cannot delete environment variable … remove it from the Docker Compose file first" | correggere il valore o i flag, non cancellare |
| Servizio senza healthcheck | stato della risorsa `unknown` | ogni servizio a lunga vita ha un healthcheck (lo verifica lo script) |

## 6. Verifica

Da VPN:

```bash
curl --fail http://giobby.cloud4job.com:2001/healthz                                   # OK
curl -s -o /dev/null -w "%{http_code}\n" http://giobby.cloud4job.com:2001/api/attributes # 401
```

- Log di `migrate`: `Applied N migration(s).` e `Database schema is current.`
- Log di `giobby-backend`: `Successfully connected to PostgreSQL`, `starting HTTP server`.
- Il proxy dell'API Explorer arriva al backend di giobby e a nessun altro: sul server
  `docker exec <container api-explorer> getent hosts giobby-backend` deve dare **un solo** IP, quello
  del container `giobby-backend` della risorsa (`docker inspect` → `NetworkSettings.Networks`).
- `http://giobby.cloud4job.com:2001/` apre l'API Explorer.
- Redeploy senza modifiche: `migrate` scrive `Applied 0 migration(s).`
- Restart della risorsa da Coolify: torna `healthy`.

## 7. Aggiornamenti e migrazioni

- Una modifica allo schema è un **nuovo file** `NNN_descrizione.sql` in
  `backend/internal/db/migrations/`, con prefisso maggiore dell'ultimo. Un file già applicato non si
  modifica mai.
- Tutte le migrazioni pendenti girano in **una transazione** con advisory lock: o passano tutte o
  nessuna. Due deploy contemporanei non le applicano due volte.
- Niente rollback automatico: per annullare una modifica si scrive una nuova migrazione.
- Backup del database `giobby` come per gli altri progetti:
  ```bash
  umask 077; docker exec $PG pg_dump -U postgres -d giobby --format=custom > giobby-$(date +%F-%H%M).dump
  ```

## 8. Diagnosi

| Sintomo | Causa / azione |
|---|---|
| Deploy fallito con `required variable <NOME> is missing a value` | variabile obbligatoria vuota in Coolify (§5) |
| Il backend esce con `JWT_SECRET is a default or placeholder value` o `must be at least 32 characters` | `JWT_SECRET` non impostato correttamente in Coolify: generarne uno con `openssl rand -hex 32` |
| Build fallito con `"/swagger.json": not found` su `api-explorer` | Build secrets = BuildKit (§5.1): passare a "Standard build arguments" |
| `migrate` esce 1 con `migration checksum mismatch: NNN_….sql` | un file di migrazione già applicato è stato modificato: ripristinarlo com'era e scrivere una nuova migrazione |
| `migrate` esce 1 con `applied migration missing from code` | si è fatto il deploy di un commit più vecchio delle migrazioni già applicate: rifare il deploy di `release` aggiornato |
| `migrate` esce 1 con `apply NNN_….sql: …` | la migrazione nuova è sbagliata; la transazione è annullata, il database è come prima |
| **Il servizio è giù dopo un `migrate` fallito** | compose ha già ricreato il backend senza avviarlo (verificato in locale). Ripristino: *Redeploy* in Coolify del deploy precedente (la migrazione fallita non ha lasciato tracce), poi correggere con una nuova PR |
| `release` non si muove dopo un merge su `main` | CI rossa su `main`: guardare la run in GitHub Actions e correggere con una PR |
| Container fermo dopo un reboot con `cannot assign requested address` | sysctl del §4 mancante; `docker ps -a` e `docker start <container>` |
| Il backend non si connette (`password authentication failed`) | la variabile non coincide con la password del ruolo: cambiare la variabile non cambia la password (`ALTER ROLE … PASSWORD …`) |
| `giobby.cloud4job.com` non risolve | il client WireGuard deve avere `DNS = 10.8.0.1`; altrimenti `http://10.8.0.1:2001` |

## 9. Fuori da questo setup

- Spegnimento del vecchio server di produzione (deploy SSH) e rimozione dei segreti GitHub non più
  usati: `DOCKERHUB_USERNAME`, `DOCKERHUB_PASSWORD`, `SSH_PRIVATE_KEY`, `SSH_PORT`, `SSH_USER`,
  `PRODUCTION_SERVER_IP`.
- Ambiente Coolify per `dev`.
