#!/usr/bin/env bash
# Exits 0 only if $SHA is the merge commit of a pull request merged into BRANCH
# (.github/branch-guard.jq). Used by the branch guard workflow and by the release job,
# so that only commits that went through a PR reach production. GitHub API errors are
# retried before giving up.
#   REPO=owner/repo SHA=<commit> GH_TOKEN=… .github/branch-guard.sh BRANCH
set -uo pipefail
branch=${1:?branch}
repo=${REPO:?REPO}
sha=${SHA:?SHA}
dir=$(cd "$(dirname "$0")" && pwd)

for attempt in 1 2 3; do
  if out=$(gh api "repos/$repo/commits/$sha/pulls") \
     && merged=$(printf '%s' "$out" | jq --arg branch "$branch" --arg sha "$sha" -f "$dir/branch-guard.jq") \
     && [ "$merged" -gt 0 ]; then
    echo "Commit $sha è il merge di una pull request in $branch."
    exit 0
  fi
  [ "$attempt" -lt 3 ] && sleep "${GUARD_RETRY_DELAY:-10}"
done
echo "::error::Il commit $sha non è il merge di una pull request in $branch (o GitHub non risponde). Vedi .github/CONTRIBUTING.md."
exit 1
