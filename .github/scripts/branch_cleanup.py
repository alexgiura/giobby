"""Deletes work branches once their PR is merged and the issues it references are closed.

Run by .github/workflows/branch-cleanup.yml when an issue or a PR is closed. An issue is
closed only by the frontend after the test in production, so the branch stays available
until then. A branch is deleted only if:
  - a PR from it was merged and none is still open;
  - it has no commits after the last merged PR (nothing would be lost);
  - every issue referenced with "Refs #n" in its merged PRs is closed (a merged PR
    without references has nothing to wait for).
dev, main and release are never touched. Needs GITHUB_TOKEN (contents: write).
DRY_RUN=1 prints the decisions without deleting anything.
"""
import json
import os
import re
import sys
import urllib.parse
import urllib.request

LONG_LIVED = {"dev", "main", "release"}
REFS = re.compile(r"\brefs\s+((?:#\d+[\s,]*)+)", re.IGNORECASE)


def referenced_issues(body):
    """Issue numbers referenced with "Refs #n" (also "Refs #8, #9")."""
    issues = set()
    for match in REFS.finditer(body or ""):
        issues.update(int(n) for n in re.findall(r"#(\d+)", match.group(1)))
    return issues


def decide(branch, head_sha, prs, issue_states):
    """(delete, reason) for one branch, given its PRs and the state of the referenced issues."""
    if branch in LONG_LIVED:
        return False, "branch permanente"
    if any(p["state"] == "open" for p in prs):
        return False, "PR ancora aperta"
    merged = [p for p in prs if p["merged"]]
    if not merged:
        return False, "nessuna PR mergiata"
    if head_sha not in {p["head_sha"] for p in merged}:
        return False, "commit dopo il merge"
    refs = set().union(*(referenced_issues(p["body"]) for p in merged))
    still_open = sorted(n for n in refs if issue_states.get(n) != "closed")
    if still_open:
        return False, "issue ancora aperte: " + ", ".join(f"#{n}" for n in still_open)
    return True, "mergiato, issue chiuse" if refs else "mergiato, nessuna issue citata"


def api(token, method, path):
    req = urllib.request.Request(
        "https://api.github.com" + path,
        method=method,
        headers={"Authorization": f"Bearer {token}", "Accept": "application/vnd.github+json"},
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.load(resp) if resp.status != 204 else None


def paged(token, path):
    page = 1
    while True:
        sep = "&" if "?" in path else "?"
        items = api(token, "GET", f"{path}{sep}per_page=100&page={page}")
        yield from items
        if len(items) < 100:
            return
        page += 1


def main():
    token, repo = os.environ["GITHUB_TOKEN"], os.environ["GITHUB_REPOSITORY"]
    owner = repo.split("/")[0]
    dry_run = os.environ.get("DRY_RUN") == "1"
    issue_states, deleted = {}, 0
    for b in paged(token, f"/repos/{repo}/branches"):
        name = b["name"]
        if name in LONG_LIVED:
            continue
        head = urllib.parse.quote(f"{owner}:{name}")
        prs = [
            {"number": p["number"], "state": p["state"], "merged": bool(p["merged_at"]),
             "head_sha": p["head"]["sha"], "body": p["body"]}
            for p in paged(token, f"/repos/{repo}/pulls?state=all&head={head}")
        ]
        for n in set().union(*(referenced_issues(p["body"]) for p in prs if p["merged"])):
            if n not in issue_states:
                issue_states[n] = api(token, "GET", f"/repos/{repo}/issues/{n}")["state"]
        delete, reason = decide(name, b["commit"]["sha"], prs, issue_states)
        print(f"{name}: {'elimino' if delete else 'tengo'} ({reason})")
        if delete and not dry_run:
            api(token, "DELETE", f"/repos/{repo}/git/refs/heads/{urllib.parse.quote(name)}")
            deleted += 1
    print(f"Branch eliminati: {deleted}")


if __name__ == "__main__":
    sys.exit(main())
