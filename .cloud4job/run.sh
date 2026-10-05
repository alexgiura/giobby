#!/usr/bin/env bash
# Single entry point for the stack's Docker-only commands, used by developers, agents and CI:
#   .cloud4job/run.sh fmt-check | lint | test | build
# Commands come from .cloud4job/stack.env (overridden by the project's optional
# .cloud4job/stack.local.env) and run from the project root with
# PROJECT_ROOT, APP_DIR and PROJECT_NAME exported.
set -euo pipefail
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"
set -a
# shellcheck disable=SC1091
. .cloud4job/project.env
# shellcheck disable=SC1091
. .cloud4job/stack.env
# Project-owned overrides (e.g. a build from the repository root): never touched by the kit.
if [ -f .cloud4job/stack.local.env ]; then
  # shellcheck disable=SC1091
  . .cloud4job/stack.local.env
fi
set +a
export PROJECT_ROOT APP_DIR="${APP_DIR:-.}"

case "${1:-}" in
  fmt-check) cmd=${FMT_CHECK_CMD:-} ;;
  lint) cmd=${LINT_CMD:-} ;;
  test) cmd=${TEST_CMD:-} ;;
  build) cmd=${BUILD_CMD:-} ;;
  *) echo "Uso: .cloud4job/run.sh fmt-check|lint|test|build" >&2; exit 2 ;;
esac
[ -n "$cmd" ] || { echo "Comando '$1' non definito per lo stack $STACK" >&2; exit 2; }
exec bash -c "$cmd"
