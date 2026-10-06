---
name: open-pr
description: Apre una pull request con la descrizione standard Cloud4Job — verso dev dai branch di lavoro, verso main da dev — solo dopo conferma esplicita.
---

# /open-pr

1. **Base**: se il branch corrente è `dev` la PR è un rilascio (`--base main`); per ogni altro
   branch di lavoro `--base dev`. Mai PR da o verso `release`.
2. **Prerequisiti**: branch già pushato (altrimenti usa `/push-check`), nessuna PR aperta per lo
   stesso branch (`gh pr list --head <branch>`).
3. **Issue**: chiedi o ricava dal nome del branch l'issue collegata.
4. **Descrizione** (in italiano):
   - **Cosa cambia** — elenco puntato dei cambiamenti;
   - **Perché** — con `Refs #<n>`, mai `Closes #<n>`: la issue la chiude il frontend dopo il test
     in produzione (vedi la sezione Progetto di `CLAUDE.md`);
   - **Verifiche fatte** — comandi lanciati e esito; cosa non è stato verificato;
   - **Punti da guardare** — rischi, scelte da validare, effetti sul deploy.
   Nessuna attribuzione AI (niente "Generated with", niente `Co-Authored-By`).
5. **Conferma**: mostra titolo, base e descrizione e chiedi conferma esplicita. Solo dopo un sì:
   `gh pr create --base dev|main --head <branch> --title … --body-file … --assignee @me`.
6. Il merge non lo fa l'agente: richiede CI verde e la review di un'altra persona.
