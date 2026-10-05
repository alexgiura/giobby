# Sviluppo con agenti AI

Come si usa Claude Code su giobby. Il repository è inizializzato con il kit AI-driven
Cloud4Job 0.1.1 (profilo medium-superpowers, stack go).

## Prerequisiti (una volta per computer)

| Strumento | Perché | Verifica |
|---|---|---|
| Git | versionamento | `git --version` |
| Docker (Desktop su macOS/Windows) | app, database e strumenti girano in container | `docker version` |
| WSL2 (solo Windows) | shell per hook e script; Docker Desktop lo usa già | `wsl --status` |
| Claude Code | agente di sviluppo | `claude --version` |
| Plugin **Superpowers** | metodo brainstorming → spec → piano → esecuzione → review | in Claude Code: `/plugin` |
| `gh` autenticato | PR, issue, CI | `gh auth status` |
| `jq` | usato dagli hook: **senza `jq` l'hook di sicurezza blocca tutto** | `jq --version` |
| Plugin Codex (facoltativo) | seconda revisione indipendente | [openai/codex-plugin-cc](https://github.com/openai/codex-plugin-cc) |

Su Windows si lavora **dentro WSL2**: si clona il repository nel filesystem Linux (`~/…`, non
`/mnt/c/…`) e si avvia Claude Code da lì.

## Come è organizzato

| Dove | Cosa |
|---|---|
| `CLAUDE.md` | regole per l'agente: blocchi del kit (comuni, profilo, stack) + sezione **Progetto** |
| `.claude/settings.json` | permessi: cosa l'agente fa da solo, cosa chiede, cosa non fa mai; attribuzione AI disattivata |
| `.claude/hooks/` | blocco dei file sensibili per gli strumenti di file e per i comandi di shell (fail-closed senza `jq`), registro delle azioni (`.claude/audit.log`), formattazione automatica dei file scritti (in Docker), log di inizio sessione |
| `.claude/skills/` | `/commit`, `/push-check`, `/open-pr` |
| `.cloud4job/` | `project.env` (configurazione del progetto), `stack.env` (comandi Docker dello stack), `stack.local.env` facoltativo (comandi del progetto che sostituiscono quelli dello stack, mai toccato dal kit), manifest dei file del kit |
| `.githooks/pre-push` | blocca i push diretti su `dev`, `main`, `release` |

## Ciclo di lavoro di un'issue

1. Prendi l'issue, spostala in **In progress**, crea il branch da `dev` (`feat/…`).
2. Lavoro nuovo: Superpowers parte dal brainstorming; spec e piano in `docs/superpowers/`.
   Correzioni piccole: si parte dall'issue.
3. L'agente implementa con test e lancia le verifiche dello stack.
4. `/commit` → review del diff, controllo segreti, commit dopo conferma.
5. `/push-check` → verifiche e push del branch dopo conferma.
6. `/open-pr` → PR verso `dev` collegata all'issue.
7. Revisione: CI verde + review umana (facoltativa una seconda revisione con Codex). Il merge lo fa
   lo sviluppatore.

## Permessi in sintesi

- **Da solo**: leggere il codice con gli strumenti di Claude Code, modificare sorgenti e test,
  comandi Git di sola lettura, `.cloud4job/run.sh` e `docker compose up|build|ps|logs|down`.
- **Chiede**: commit, push, PR, issue, `gh api`, eliminazioni e spostamenti, rebase/reset/restore,
  `docker run` e `docker compose run|exec`, file di CI, `.claude/`, Dockerfile, compose,
  dipendenze, migrazioni.
- **Mai**: leggere `.env*` (tranne `.env.example`), chiavi, keystore, `~/.ssh`, `~/.config/gh`;
  push forzati o su `dev`/`main`/`release`; `sudo`; script scaricati ed eseguiti al volo; gestire
  segreti o autenticazione di `gh`.

Come è garantito: le regole `deny` valgono per gli strumenti di file di Claude Code; per la shell
c'è l'hook `check-sensitive-command.sh`, che blocca i comandi che citano file con segreti o che
danno a un container accesso al computer (mount di `/` o della home, socket Docker,
`--privileged`). Sono controlli a pattern: riducono gli errori, non sostituiscono la revisione di
quello che l'agente propone di eseguire. In caso di dubbio l'agente chiede.

Eccezioni personali in `.claude/settings.local.json` (non committato), non nel file condiviso.

**Prima apertura del repository**: Claude Code chiede se la cartella è attendibile; permessi e hook
del repository si applicano dopo l'accettazione. Accettare una volta, dopo aver verificato di essere
nel repository giusto.

## Aggiornare il kit

Dal repository del kit, alla versione nuova:

```bash
./init-ai-template.sh --update --target <percorso del progetto> --dry-run
./init-ai-template.sh --update --target <percorso del progetto>
```

I file del kit non modificati nel progetto si aggiornano; quelli modificati producono un
`FILE.conflict` da unire a mano. La sezione **Progetto** di `CLAUDE.md` e `.cloud4job/project.env`
non vengono mai toccati. Il risultato entra con una normale PR.
