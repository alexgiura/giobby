#!/bin/sh
# Tests the branch guard filter (.github/branch-guard.jq) on sample responses of
# GET /repos/{repo}/commits/{sha}/pulls: a push is accepted only if SHA is the
# merge commit of a pull request merged into that same branch.
set -eu
cd "$(dirname "$0")/.."
FILTER=.github/branch-guard.jq
SHA=abc123

count() { echo "$1" | jq --arg branch "$2" --arg sha "$SHA" -f "$FILTER"; }
expect() {
  got=$(count "$2" "$3")
  [ "$got" = "$4" ] || { echo "FAIL: $1: got $got, want $4" >&2; exit 1; }
}

merged_into_dev='[{"merged_at":"2026-10-02T10:00:00Z","base":{"ref":"dev"},"merge_commit_sha":"abc123"}]'
merged_elsewhere_commit='[{"merged_at":"2026-10-02T10:00:00Z","base":{"ref":"dev"},"merge_commit_sha":"zzz999"}]'
open_pr='[{"merged_at":null,"base":{"ref":"main"},"merge_commit_sha":"abc123"}]'

expect "PR merged into dev, push on dev" "$merged_into_dev" dev 1
expect "PR merged into dev, same commit pushed to main (dev:main)" "$merged_into_dev" main 0
expect "commit belongs to a PR but is not its merge commit" "$merged_elsewhere_commit" dev 0
expect "open PR only" "$open_pr" main 0
expect "no PR" '[]' main 0

echo "branch guard filter OK"
