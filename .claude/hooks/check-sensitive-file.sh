#!/usr/bin/env bash
# PreToolUse hook (Read, Edit, Write, Grep, Glob…): blocks paths that may hold secrets.
# Exit 2 blocks the call and shows the message to the agent. Fail-closed: without jq
# the payload cannot be checked, so every call is blocked.
set -u

if ! command -v jq >/dev/null 2>&1; then
  echo "Hook di sicurezza: jq non è installato, operazione bloccata. Installare jq (vedi docs/AI.md)." >&2
  exit 2
fi

input=$(cat)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty')
[ -n "$cwd" ] || cwd=${CLAUDE_PROJECT_DIR:-$PWD}

# normalize PATH: absolute path with ~, ., .. resolved (the file may not exist yet).
normalize() {
  local p=$1 out="" part
  # A literal "~" written by the agent means the home directory.
  # shellcheck disable=SC2088
  case "$p" in "~"|"~/"*) p="$HOME${p#\~}" ;; /*) ;; *) p="$cwd/$p" ;; esac
  local IFS=/
  set -f  # path parts may contain glob characters: never expand them against the disk
  for part in $p; do
    case "$part" in
      ""|.) ;;
      ..) out=${out%/*} ;;
      *) out="$out/$part" ;;
    esac
  done
  set +f
  printf '%s' "${out:-/}"
}

# is_sensitive PATH: true when the path matches a secret-holding pattern (case-insensitive).
is_sensitive() {
  local lower base
  lower=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')
  base=${lower##*/}
  case "/$lower/" in */.ssh/*|*/.config/gh/*|*/.aws/*|*/.local/*) return 0 ;; esac
  case "$base" in
    .env.example|.env.sample|.env.template|.env.dist) return 1 ;;
    .env*) return 0 ;;
    *.pem|*.key|*.p12|*.pfx|*.jks|*.keystore) return 0 ;;
    id_rsa*|id_dsa*|id_ecdsa*|id_ed25519*) return 0 ;;
    local.properties) return 0 ;;
  esac
  [[ "$base" =~ ^(credentials?|secrets?)(\.(json|ya?ml|xml|txt|env|ini))?$ ]]
}

deny() {
  echo "Accesso bloccato a '$1': file che può contenere segreti (regole Cloud4Job, .claude/hooks/check-sensitive-file.sh). Se serve davvero, chiedi allo sviluppatore." >&2
  exit 2
}

# Glob patterns could match env files however they are written ({.env,x}, .[e]nv):
# strip the glob syntax and look for an env name that is not an example file.
glob_may_match_env() {
  local flat
  flat=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | tr -d '[]{}*?')
  flat=$(printf '%s' "$flat" | sed -E 's/\.env\.(example|sample|template|dist)//g')
  case "$flat" in *.env*) return 0 ;; esac
  return 1
}

tool=$(printf '%s' "$input" | jq -r '.tool_name // empty')

# Paths: file_path / path / notebook_path, for every file tool.
while IFS= read -r candidate; do
  [ -n "$candidate" ] || continue
  is_sensitive "$(normalize "$candidate")" && deny "$candidate"
done < <(printf '%s' "$input" | jq -r '.tool_input // {} | [.file_path, .path, .notebook_path] | map(select(type == "string" and . != "")) | .[]')

# Globs: Glob's pattern and Grep's glob filter. Grep's pattern is a content regex, not a path.
while IFS= read -r glob; do
  [ -n "$glob" ] || continue
  if glob_may_match_env "$glob" || is_sensitive "$(normalize "$glob")"; then deny "$glob"; fi
done < <(printf '%s' "$input" | jq -r --arg t "$tool" '.tool_input // {} | [(if $t == "Glob" then .pattern else null end), .glob] | map(select(type == "string" and . != "")) | .[]')

exit 0
