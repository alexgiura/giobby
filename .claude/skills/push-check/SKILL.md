---
name: push-check
description: Verifica che un branch sia pronto e lo pusha su GitHub solo dopo conferma esplicita — mai su dev, main o release, mai forzato. Usala prima di ogni push.
---

# /push-check

1. **Branch**: `git branch --show-current`. Rifiuta il push se il branch è `dev`, `main` o `release`:
   su `dev` e `main` si arriva solo con una PR, `release` lo aggiorna solo la CI.
2. **Stato**: `git status` pulito (niente modifiche non committate, niente `*.conflict`).
3. **Base aggiornata**: `git fetch origin` e controlla se il branch è indietro rispetto a
   `origin/dev`. Se lo è, proponi `git pull --rebase origin dev` (chiede conferma) e rilancia le
   verifiche.
4. **Verifiche**: formattazione, lint e test dello stack (`.cloud4job/stack.env`). Se falliscono,
   fermati e riporta l'errore: non si pusha codice rosso.
5. **Riepilogo**: commit che verranno pubblicati (`git log --oneline origin/dev..HEAD`).
6. **Conferma**: chiedi conferma esplicita. Solo dopo un sì: `git push -u origin <branch>`.
7. Mai `--force`, mai `--no-verify`. Se il push viene rifiutato dall'hook, non aggirarlo: riporta il
   motivo.
