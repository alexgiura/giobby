#!/usr/bin/env bash
# PreToolUse hook (Bash): blocks shell commands that would read secret files or give a
# container access to the host (the Read deny rules do not cover shell subprocesses).
# Exit 2 blocks the call. Fail-closed: without jq every command is blocked.
set -u

if ! command -v jq >/dev/null 2>&1; then
  echo "Hook di sicurezza: jq non è installato, comando bloccato. Installare jq (vedi docs/AI.md)." >&2
  exit 2
fi

cmd=$(jq -r '.tool_input.command // empty' | tr '[:upper:]' '[:lower:]')
[ -n "$cmd" ] || exit 0

block() {
  echo "Comando bloccato ($1): può esporre segreti o il computer dello sviluppatore (regole Cloud4Job, .claude/hooks/check-sensitive-command.sh). Se serve davvero, chiedi allo sviluppatore di eseguirlo." >&2
  exit 2
}

# Env files: any .env or .env.<name> except the example variants.
while IFS= read -r token; do
  case "$token" in
    .env.example|.env.sample|.env.template|.env.dist) ;;
    *) block "file $token" ;;
  esac
done < <(printf '%s' "$cmd" | grep -oE '\.env(\.[a-z0-9_-]+)?' || true)

case "$cmd" in
  *.ssh/*|*id_rsa*|*id_dsa*|*id_ecdsa*|*id_ed25519*|*.config/gh*|*.aws/*|*.pem*|*.p12*|*.pfx*|*.jks*|*.keystore*|*local.properties*)
    block "chiavi o credenziali" ;;
esac

# Scripts downloaded and executed on the fly.
if printf '%s' "$cmd" | grep -qE '(curl|wget)[^|]*\|[[:space:]]*(sudo[[:space:]]+)?(sh|bash|zsh)([[:space:]]|$)'; then
  block "script scaricato ed eseguito al volo"
fi

# git push: never to dev/main/release (any refspec form), never forced, never without hooks.
push_args=$(printf '%s' "$cmd" | grep -oE 'git[[:space:]]+push[^;&|]*' | head -n 1 | sed -E 's/^git[[:space:]]+push//') || true
if [ -n "$push_args" ] || printf '%s' "$cmd" | grep -qE 'git[[:space:]]+push([[:space:]]|$)'; then
  remote_seen=false
  for arg in $push_args; do
    case "$arg" in
      -f|--force|--force-with-lease*|--force-if-includes|--mirror|--delete|-d) block "git push $arg" ;;
      --no-verify) block "git push --no-verify" ;;
      -*) continue ;;
    esac
    if [ "$remote_seen" = false ]; then remote_seen=true; continue; fi
    dest=${arg##*:}; dest=${dest#+}; dest=${dest#refs/heads/}
    case "$dest" in main|dev|release) block "push su $dest: si arriva solo con una pull request" ;; esac
  done
fi

# Containers that see the host: root or home mounts, the Docker socket, privileged mode.
case "$cmd" in
  *docker*)
    # The literal text "$home" (the command was lowercased) is what we look for.
    # shellcheck disable=SC2016
    case "$cmd" in
      *--privileged*|*--pid=host*|*docker.sock*|*"-v /:"*|*"-v \"/:"*|*"--volume /:"*|*"--volume=/:"*|*"-v ~"*|*'-v $home'*|*'-v "$home'*|*'-v ${home}'*|*'-v "${home}'*)
        block "container con accesso al computer" ;;
    esac ;;
esac

exit 0
