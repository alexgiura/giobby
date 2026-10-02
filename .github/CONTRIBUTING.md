# Come contribuire a giobby-backend

Regole riprese dallo starter kit di try2catch
([Cloud4Job/try2catch-me-docs → CONTRIBUTING.md](https://github.com/Cloud4Job/try2catch-me-docs/blob/main/CONTRIBUTING.md)),
adattate al flusso di giobby.

## Preparazione (una volta per clone)

```bash
git config core.hooksPath .githooks
cp backend/.env.example backend/.env
```

L'hook `.githooks/pre-push` blocca i push diretti su `dev`, `main` e `release`.

## Flusso dei branch

```
feat/… fix/… chore/… docs/…  ──PR──▶  dev  ──PR──▶  main  ──CI verde──▶  release  ──▶  produzione (Coolify)
```

1. **Branch** da `dev` aggiornato: `git switch dev && git pull --ff-only && git switch -c feat/<cosa>`.
2. **Commit** piccoli e coerenti, messaggio all'imperativo che spiega il perché.
   Nessuna attribuzione AI (`Co-Authored-By` di agenti, "Generated with…").
3. **Verifiche locali**: `docker compose up -d --build`, test Go (vedi la CI in `.github/workflows/ci.yml`).
4. **Push del branch** (mai di `dev`, `main` o `release`) e **pull request verso `dev`**: cosa cambia, perché, verifiche fatte, punti da guardare.
5. **CI verde** e **review** di un'altra persona, poi merge dalla pagina della PR.
6. **Rilascio**: pull request `dev` → `main`. Dopo il merge la CI gira su `main` e, solo se è verde,
   porta `release` al commit di `main`; Coolify fa il deploy da `release`.

## `dev`, `main`, `release`

- Su `dev` e `main` si arriva **solo** con una PR. Il workflow **Guardia branch** segnala in rosso
  ogni push che non arriva da una PR unita (con GitHub Free la protezione dei branch non è disponibile).
- `release` lo aggiorna **solo** la CI. Non si tocca a mano: se `release` resta indietro, la CI su
  `main` è rossa e va sistemata con una nuova PR.
- Mai `git push --force` su questi branch, mai `--no-verify` per aggirare l'hook.

## Database

Lo schema è nelle migrazioni in `backend/internal/db/migrations/`. Una modifica allo schema è un
**nuovo file** con prefisso numerico maggiore dell'ultimo; un file già applicato non si modifica mai
(il servizio `migrate` si ferma con `checksum mismatch`).

## Sicurezza

- Nessun segreto in codice, test, log, issue o messaggi: in locale solo valori sintetici
  (`backend/.env`, escluso da git). I valori di produzione stanno solo in Coolify.
- Un segreto finito per errore in un commit va revocato, non solo cancellato dalla storia.
