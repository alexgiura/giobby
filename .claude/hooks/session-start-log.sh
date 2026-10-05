#!/usr/bin/env bash
# SessionStart hook: records the session with profile, stack and kit version.
set -u
command -v jq >/dev/null 2>&1 || exit 0
project=${CLAUDE_PROJECT_DIR:-$PWD}
input=$(cat)
value() { grep -E "^$1=" "$project/.cloud4job/project.env" 2>/dev/null | tail -n 1 | cut -d= -f2-; }
mkdir -p "$project/.claude"
jq -cn --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg session "$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null)" \
  --arg stack "$(value STACK)" --arg kit "$(value KIT_VERSION)" \
  '{ts: $ts, event: "session_start", session: $session, profile: "medium-superpowers", stack: $stack, kit_version: $kit}' \
  >> "$project/.claude/audit.log"
exit 0
