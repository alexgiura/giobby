---
name: commit
description: Prepara un commit secondo le regole Cloud4Job — rivede il diff, cerca segreti, propone il messaggio e committa solo dopo conferma esplicita. Usala ogni volta che c'è da committare.
---

# /commit

1. **Branch**: `git branch --show-current`. Se è `dev`, `main` o `release`, fermati: si lavora solo
   su un branch `feat/…`, `fix/…`, `chore/…`, `docs/…` (proponi di crearlo da `dev`).
2. **Cosa entra**: `git status` e `git diff` (più `git diff --cached` se c'è già qualcosa in stage).
   Escludi file non correlati al task e file generati; segnala i file `*.conflict` del kit (vanno
   uniti a mano, mai committati).
3. **Segreti**: cerca nel diff chiavi, token, password, connection string con credenziali, file
   `.env` (tranne `.env.example`), chiavi private. Se trovi qualcosa fermati e segnalalo: va tolto e,
   se era reale, ruotato.
4. **Verifiche**: lancia i comandi di formattazione, lint e test dello stack
   (`.cloud4job/stack.env`, sezione stack di `CLAUDE.md`). Se falliscono, riportalo prima di
   proporre il commit.
5. **Messaggio**: titolo all'imperativo (max ~72 caratteri) che dice cosa fa il cambiamento; corpo,
   se serve, con il perché. Lingua: quella dei commit esistenti (`git log --oneline -10`).
   **Nessuna attribuzione AI**: niente `Co-Authored-By` di agenti, niente "Generated with".
6. **Conferma**: mostra l'elenco dei file e il messaggio, e chiedi conferma esplicita. Solo dopo un
   sì: `git add <file…>` e `git commit`.
7. Mai `--no-verify`, mai `--amend` su commit già pushati senza richiesta esplicita.
