# giobby — istruzioni per l'agente

Kit Cloud4Job 0.1.1 · profilo medium-superpowers · stack go

<!-- kit:common -->
## Regole comuni Cloud4Job

Valgono per ogni repository che usa il kit e per ogni agente (Claude Code, Codex o altri). Le regole
del profilo e dello stack, più sotto, le specializzano; la sezione **Progetto** in fondo aggiunge
quelle del singolo repository. In caso di conflitto prevale la regola più restrittiva.

### Prima di modificare

1. Leggi questo file, la sezione Progetto e l'issue di riferimento.
2. Se requisito o criterio di accettazione non sono chiari, chiedi. Non inventare requisiti né
   comportamenti non descritti.
3. Individua i file coinvolti e limita il lavoro a quelli.

### Perimetro

- Tocca solo ciò che serve al task. Niente refactoring, pulizie o "miglioramenti" non richiesti: se
  noti qualcosa, segnalalo o proponi una nuova issue.
- Non modificare build, CI/CD, dipendenze o infrastruttura senza richiesta esplicita e motivata.
- Le decisioni tecniche rilevanti emerse durante il lavoro vanno sottoposte allo sviluppatore, non
  risolte con una scelta arbitraria.

### Tutto in Docker

- Applicazione, database e strumenti (build, test, lint, formattazione) girano in container.
- Usa **solo** i comandi della sezione dello stack (`.cloud4job/stack.env`) o del `docker compose`
  del progetto. Non usare toolchain installate sul computer (`go`, `node`, `python`, `flutter`…)
  e non chiedere di installarle.
- Eccezione dichiarata dallo stack mobile: emulatori e build iOS.

### Sicurezza

- Non leggere file con segreti o configurazioni private: `.env*` (tranne `.env.example`),
  `.local/`, chiavi, keystore, file con password/secret/token/credential nel nome. Gli hook lo
  bloccano comunque.
- Non scrivere mai credenziali, token o chiavi in codice, configurazioni, test, commenti, log,
  messaggi di commit o PR. Solo valori **sintetici** di sviluppo.
- I valori di produzione stanno solo nel pannello di deploy (Coolify), mai nel repository.
- Un segreto finito per errore in un commit, in un log o in un'immagine va **ruotato**, non solo
  cancellato: segnalalo subito.

### Operazioni che richiedono conferma esplicita

Prima di eseguirle chiedi e attendi un sì:

- `git commit`, `git push` (sempre su branch, mai forzato), apertura o merge di PR, issue;
- comandi che possono perdere lavoro locale (`reset`, `rebase`, `restore`, `stash drop`);
- eliminazione di file, branch, dati, volumi o database;
- modifiche a `.github/`, `.claude/`, `.githooks/`, `.cloud4job/`, Dockerfile, compose, dipendenze,
  migrazioni;
- qualsiasi azione verso servizi esterni o non reversibile (server, pannelli, API).

### Git

- Branch `feat/…`, `fix/…`, `chore/…`, `docs/…` creati da `dev` aggiornato; si integra solo con PR
  verso `dev`, con CI verde e review umana. Il rilascio è una PR `dev` → `main`.
- `release` si aggiorna solo dalla CI dopo una CI verde su `main`: non toccarlo mai.
- Commit piccoli e coerenti, messaggio all'imperativo che spiega il perché.
- **Nessuna attribuzione AI**: niente `Co-Authored-By` di agenti, niente "Generated with…".
- Prima del commit: rivedi il diff, verifica che non ci siano segreti, chiedi conferma.
- Mai `--no-verify`, mai `push --force`.

### Qualità

- Scrivi come il codice circostante: stile, densità dei commenti, nomi, idiomi. Commenta solo il
  perché non ovvio.
- Logica verificabile → test. Correzione di un bug → prima un test che lo riproduce.
- Esegui le verifiche dello stack (formattazione, lint, test) prima di dichiarare finito.

### Feedback di revisione

Che arrivi da una persona, da Codex o da un altro strumento: valuta ogni segnalazione nel merito,
applica quelle fondate e **riporta esplicitamente** quelle che non accogli, con il motivo. Non
recepire in modo automatico e non scartare in silenzio.

### Chiusura del task

Concludi sempre con un riepilogo: cosa è cambiato, perché, in quali file; verifiche eseguite e
relativo esito (se qualcosa non è stato eseguito, dillo); limiti, ipotesi e punti aperti; passi
successivi o nuove issue da aprire. Non dichiarare che qualcosa funziona se non è stato eseguito.
<!-- /kit:common -->

<!-- kit:profile -->
## Profilo: medium-superpowers

Il repository usa il metodo **Superpowers** (plugin di Claude Code): brainstorming → spec → piano →
esecuzione → review. Il plugin deve essere installato e attivo nell'ambiente dello sviluppatore
(`/plugin` in Claude Code). La scelta del profilo è una decisione umana: non cambiarla.

- **Lavoro nuovo** (funzionalità, cambi di comportamento, nuovi componenti): parte dalla skill di
  brainstorming; spec e piano approvati dallo sviluppatore prima di scrivere codice, salvati in
  `docs/superpowers/specs/` e `docs/superpowers/plans/` e committati con il lavoro.
- **Correzioni piccole e localizzate** (bug con causa chiara, testi, configurazioni): basta l'issue
  con il criterio di accettazione; resta valido il test che riproduce il bug.
- **Esecuzione**: test prima del codice (TDD), una verifica a ogni passo, review indipendente prima
  del merge. Non dichiarare finito un passo senza averne visto l'esito.
- **Debug**: con la skill di debugging sistematico; prima la causa, poi la correzione.
- **Seconda revisione facoltativa**: il plugin Codex per Claude Code può rivedere piani e codice; le
  sue osservazioni seguono le regole sul feedback di revisione.
<!-- /kit:profile -->

<!-- kit:stack -->
## Stack: go

- **Comandi** (sempre questi, mai un `go` installato sul computer):
  `.cloud4job/run.sh fmt-check` · `lint` (`go vet`) · `test` (`go test ./...`) · `build` (immagine
  Docker). Per comandi Go occasionali usa lo stesso container:
  `docker run --rm -v "$PWD":/src -w /src golang:1.22.12-alpine go <comando>`.
- **Formattazione**: `gofmt`, applicato dall'hook dopo ogni modifica a un `.go`; la CI fallisce su
  file non formattati.
- **Dipendenze**: `go.mod`/`go.sum` si cambiano solo con richiesta motivata (chiede conferma).
- **HTTP**: l'app espone `GET /healthz` senza autenticazione (`200 OK`), usato dagli healthcheck.
- **Configurazione** da variabili d'ambiente. In produzione (`APP_ENV=production` o equivalente)
  l'app **rifiuta di partire** con segreti di default, segnaposto (`missing…`) o troppo corti.
- **Database**: migrazioni versionate e immutabili, eseguite da un comando separato prima dell'app
  (mai all'avvio dell'API); l'API usa un ruolo senza DDL. Test di integrazione su Postgres alla
  stessa versione della produzione.
- **Errori**: verso il client messaggi generici, dettagli solo nei log, mai segreti nei log.
<!-- /kit:stack -->

<!-- project -->
## Progetto

API Go del gestionale **Giobby** (aziende, anagrafiche, magazzino, documenti di vendita e acquisto,
contabilità, CRM, calendario, messaggi), con PostgreSQL e un API Explorer. Produzione su Coolify, solo
in VPN: `http://giobby.cloud4job.com:2001` (guida: `docs/DEPLOY-COOLIFY.md`).

### Mappa del codice

| Percorso | Contenuto |
|---|---|
| `backend/cmd/main.go` | avvio dell'API HTTP |
| `backend/cmd/migrate` | comando delle migrazioni (servizio `migrate`, prima dell'API a ogni deploy) |
| `backend/internal/db/migrations/` | migrazioni SQL versionate (`NNN_nome.sql`) |
| `backend/internal/db/migrate.go` | runner: transazione, advisory lock, checksum SHA-256, permessi al ruolo app |
| `backend/internal/config` | configurazione da variabili d'ambiente; in produzione rifiuta `JWT_SECRET` deboli |
| `backend/internal/routes`, `handlers`, `services`, `repository`, `models` | API: rotte → handler → servizi → accesso ai dati |
| `backend/internal/auth`, `middleware` | JWT, sessioni, autenticazione delle richieste |
| `backend/internal/openapi`, `swagger.json` | documento OpenAPI servito dall'API |
| `api-explorer/` | nginx con l'interfaccia per provare le API e proxy `/api/` verso `giobby-backend` |
| `docker-compose.yml` | stack locale: Postgres 18.6 con ruoli di sviluppo, `migrate`, API, API Explorer |
| `docker-compose.prod.yml` | produzione su Coolify (porta `10.8.0.1:2001`) |

### Comandi

```bash
cp backend/.env.example backend/.env      # una volta, solo valori di sviluppo
docker compose up -d --build              # API su 127.0.0.1:8080, API Explorer su 127.0.0.1:3000
.cloud4job/run.sh fmt-check | lint | test | build
scripts/check-prod-compose.sh             # controllo del compose di produzione
```

`test` avvia un PostgreSQL 18.6 usa e getta e passa ai test `GIOBBY_TEST_DATABASE_URL`
(comandi del progetto in `.cloud4job/stack.local.env`).

### Regole del progetto

- **Migrazioni**: mai modificare un file già applicato; ogni cambio di schema è un nuovo file con
  prefisso maggiore dell'ultimo. Le migrazioni non partono all'avvio dell'API.
- **Ruoli del database**: `giobby` (proprietario, migrazioni) e `giobby_app` (API, solo dati, niente
  DDL). Nuove tabelle e sequenze ricevono i permessi da `migrate`.
- **Contratto HTTP**: ogni cambio di API aggiorna `swagger.json`; errori generici verso il client.
- **Servizi di produzione**: nomi univoci sulla rete condivisa di Coolify (`giobby-backend`); il proxy
  dell'API Explorer usa `BACKEND_HOST`, mai un nome generico.
- **Segreti**: solo in Coolify, "Not available during build"; `backend/.env` è solo locale e
  ignorato da git.
<!-- /project -->
