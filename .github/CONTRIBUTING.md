# Come contribuire a giobby

Regole comuni dei progetti Cloud4Job (kit AI-driven 0.1.1). Per lavorare con agenti AI
vedi anche [`docs/AI.md`](../docs/AI.md).

## Preparazione (una volta per clone)

Sul computer servono solo **Git, Docker, `gh` e `jq`** (su Windows dentro WSL2). Tutto il resto
gira in container.

```bash
git config core.hooksPath .githooks   # blocca i push diretti su dev, main, release
cp backend/.env.example backend/.env  # solo valori di sviluppo
docker compose up -d --build
```

## Flusso dei branch

```
feat/… fix/… chore/… docs/…  ──PR──▶  dev  ──PR──▶  main  ──CI verde──▶  release  ──▶  deploy
```

1. Ogni lavoro parte da un'**issue**. Branch da `dev` aggiornato:
   `git switch dev && git pull --ff-only && git switch -c feat/<cosa>`.
2. **Commit** piccoli e coerenti, messaggio all'imperativo che spiega il perché. Nessuna
   attribuzione AI (`Co-Authored-By` di agenti, "Generated with…").
3. **Verifiche locali** con i comandi dello stack (formattazione, lint, test) in Docker.
4. **Push del branch** (mai di `dev`, `main` o `release`) e **pull request verso `dev`**: cosa
   cambia, perché (`Closes #<n>`), verifiche fatte, punti da guardare.
5. **CI verde** e **review** di un'altra persona, poi merge dalla pagina della PR ed eliminazione
   del branch.
6. **Rilascio**: pull request `dev` → `main`. Dopo il merge la CI gira su `main` e, solo se è verde
   e il commit è il merge di una PR, porta `release` al commit di `main`; il deploy parte da `release`.

## `dev`, `main`, `release`

- Su `dev` e `main` si arriva **solo** con una PR. Il workflow **Guardia branch** segnala in rosso
  ogni push che non è il merge di una PR in quel branch (con GitHub Free la protezione dei branch
  non è disponibile sui repository privati).
- `release` lo aggiorna **solo** la CI. Se resta indietro, la CI su `main` è rossa e va sistemata
  con una nuova PR.
- Mai `git push --force` su questi branch, mai `--no-verify` per aggirare l'hook.

## Project "Giobby API"

Le attività stanno nel project https://github.com/orgs/Cloud4Job/projects/4: Backlog → Todo → In progress
→ Review → **Test** (in produzione, issue assegnata al frontend) → **Done** (il frontend ha testato e chiuso
la issue). Nei commit e nelle PR le issue si citano con **`Refs #n`**, mai `Closes #n`. "Stima (h)" si
compila a mano; "Effettivo (h)" lo calcola il workflow `project-effort.yml` (ore lavorative lun–ven 9–18,
da In progress a Test) con il segreto `PROJECT_TOKEN`.

## Database

Lo schema cambia solo con **nuove migrazioni** in `backend/internal/db/migrations/`: un nuovo file
con prefisso numerico maggiore dell'ultimo. Un file già applicato non si modifica mai (il servizio
`migrate` si ferma con `checksum mismatch`).

## Sicurezza

- Nessun segreto in codice, test, log, issue o messaggi: in locale solo valori sintetici. I valori
  di produzione stanno solo nel pannello di deploy.
- Un segreto finito per errore in un commit va **ruotato**, non solo cancellato dalla storia.
