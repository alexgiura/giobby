#!/usr/bin/env bash
# PostToolUse hook (Edit, Write): formats the written file with the stack's formatter,
# in Docker (FMT_CMD in .cloud4job/stack.env, which reads $PROJECT_ROOT and $FILE). Never blocks
# the agent: if Docker is not available or the formatter fails, it exits 1, which Claude Code
# shows to the user as a non-blocking warning (exit 2 would block).
set -u
command -v jq >/dev/null 2>&1 || exit 0
project=${CLAUDE_PROJECT_DIR:-}
stack_env="$project/.cloud4job/stack.env"
[ -n "$project" ] && [ -f "$stack_env" ] || exit 0

file=$(jq -r '.tool_input.file_path // empty')
[ -n "$file" ] || exit 0
case "$file" in /*) ;; *) file="$project/$file" ;; esac
case "$file" in "$project"/*) ;; *) exit 0 ;; esac
rel=${file#"$project"/}

FMT_CMD="" FMT_EXTENSIONS=""
# shellcheck disable=SC1090
. "$stack_env"
[ -n "$FMT_CMD" ] || exit 0
ext=${file##*.}
case " $FMT_EXTENSIONS " in *" $ext "*) ;; *) exit 0 ;; esac

# Names travel as environment variables, never spliced into the command text: a file
# name with shell syntax (e.g. "a&cmd&.go") cannot run anything.
if ! out=$(PROJECT_ROOT="$project" FILE="$rel" bash -c "$FMT_CMD" 2>&1); then
  printf 'Formattazione di %s non eseguita (Docker non attivo, immagine non ancora scaricata: lanciare .cloud4job/run.sh lint, o errore del formatter): %s\n' "$rel" "$out" >&2
  exit 1
fi
exit 0
