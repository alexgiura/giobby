#!/usr/bin/env bash
# PostToolUse hook: one JSON line per tool call in .claude/audit.log (ignored by git).
# Records which tool touched which file; never file contents, and for shell commands
# only the program name (skipping VAR=value prefixes), since arguments may carry tokens;
# URLs lose their query string.
set -u
command -v jq >/dev/null 2>&1 || exit 0
input=$(cat)
dir=${CLAUDE_PROJECT_DIR:-$(printf '%s' "$input" | jq -r '.cwd // empty')}
[ -n "$dir" ] || exit 0
mkdir -p "$dir/.claude"
printf '%s' "$input" | jq -c --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '{
  ts: $ts,
  event: (.hook_event_name // "PostToolUse"),
  session: .session_id,
  tool: .tool_name,
  target: (.tool_input // {} | .file_path // .notebook_path // .path //
           (if .command then (.command | split(" ") | map(select(length > 0))
              | map(select(test("^[A-Za-z_][A-Za-z0-9_]*=") | not)) | .[0]) else null end) //
           .pattern // (if .url then (.url | sub("[?#].*$"; "")) else null end))
}' >> "$dir/.claude/audit.log" 2>/dev/null
exit 0
