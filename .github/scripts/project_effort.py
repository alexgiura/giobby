"""Fills "Effettivo (h)" in the Giobby API project from the card's history.

Run every few minutes by .github/workflows/project-effort.yml. GitHub Projects has no
computed fields and Actions get no events when a card moves, so each run looks at every
card and records:
  - "Inizio sviluppo" (UTC ISO time) the first time the card is seen In progress;
  - "Fine sviluppo" when the card reaches Test (cleared if work resumes after a failed test);
  - "Effettivo (h)" when the card reaches Done: working hours (Mon-Fri 9-18 Europe/Rome)
    between the two, i.e. development time comparable with "Stima (h)".
Values already written are never overwritten. Needs PROJECT_TOKEN (Projects read/write).
DRY_RUN=1 prints the changes without writing them.
"""
import json
import os
import sys
import urllib.request
from datetime import datetime, time, timedelta, timezone
from zoneinfo import ZoneInfo

ROME = ZoneInfo("Europe/Rome")
DAY_START, DAY_END = time(9), time(18)

STATUS, START, END, EFFECTIVE = "Status", "Inizio sviluppo", "Fine sviluppo", "Effettivo (h)"


def working_hours(start, end):
    """Hours between two aware datetimes that fall Mon-Fri 9-18 in Europe/Rome."""
    if end <= start:
        return 0.0
    start, end = start.astimezone(ROME), end.astimezone(ROME)
    total = timedelta()
    day = start.date()
    while day <= end.date():
        if day.weekday() < 5:
            open_ = datetime.combine(day, DAY_START, ROME)
            close = datetime.combine(day, DAY_END, ROME)
            lo, hi = max(open_, start), min(close, end)
            if hi > lo:
                total += hi - lo
        day += timedelta(days=1)
    return round(total.total_seconds() / 3600, 1)


def parse(ts):
    return datetime.fromisoformat(ts.replace("Z", "+00:00"))


def decide(status, start, end, effective, now):
    """Field changes for one card: {"start"|"end": iso or None, "effective": hours}."""
    if status == "In progress":
        if not start:
            return {"start": now}
        if end:
            return {"end": None}  # back in development after a failed test
        return {}
    if status == "Test" and start and not end:
        return {"end": now}
    if status == "Done" and start and effective is None:
        changes = {} if end else {"end": now}
        changes["effective"] = working_hours(parse(start), parse(end or now))
        return changes
    return {}


QUERY = """
query($owner: String!, $number: Int!, $after: String) {
  organization(login: $owner) {
    projectV2(number: $number) {
      id
      fields(first: 50) { nodes { ... on ProjectV2FieldCommon { id name } } }
      items(first: 100, after: $after) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          fieldValues(first: 30) {
            nodes {
              ... on ProjectV2ItemFieldSingleSelectValue { name field { ... on ProjectV2FieldCommon { name } } }
              ... on ProjectV2ItemFieldTextValue { text field { ... on ProjectV2FieldCommon { name } } }
              ... on ProjectV2ItemFieldNumberValue { number field { ... on ProjectV2FieldCommon { name } } }
            }
          }
        }
      }
    }
  }
}"""

SET = """mutation($p: ID!, $i: ID!, $f: ID!, $v: ProjectV2FieldValue!) {
  updateProjectV2ItemFieldValue(input: {projectId: $p, itemId: $i, fieldId: $f, value: $v}) { clientMutationId } }"""
CLEAR = """mutation($p: ID!, $i: ID!, $f: ID!) {
  clearProjectV2ItemFieldValue(input: {projectId: $p, itemId: $i, fieldId: $f}) { clientMutationId } }"""


def graphql(token, query, variables):
    req = urllib.request.Request(
        "https://api.github.com/graphql",
        data=json.dumps({"query": query, "variables": variables}).encode(),
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=30) as resp:
        body = json.load(resp)
    if body.get("errors"):
        raise RuntimeError(body["errors"])
    return body["data"]


def card_values(item):
    values = {}
    for v in item["fieldValues"]["nodes"]:
        if not v or "field" not in v:
            continue
        name = v["field"]["name"]
        values[name] = v.get("name", v.get("text", v.get("number")))
    return values


def main():
    token = os.environ["PROJECT_TOKEN"]
    owner, number = os.environ.get("PROJECT_OWNER", "Cloud4Job"), int(os.environ.get("PROJECT_NUMBER", "4"))
    now = datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
    dry_run = os.environ.get("DRY_RUN") == "1"
    after, changed = None, 0
    while True:
        project = graphql(token, QUERY, {"owner": owner, "number": number, "after": after})["organization"]["projectV2"]
        fields = {f["name"]: f["id"] for f in project["fields"]["nodes"] if f}
        for item in project["items"]["nodes"]:
            v = card_values(item)
            changes = decide(v.get(STATUS), v.get(START), v.get(END), v.get(EFFECTIVE), now)
            for key, value in changes.items():
                field = fields[{"start": START, "end": END, "effective": EFFECTIVE}[key]]
                if dry_run:
                    print(f"[dry-run] {item['id']} {v.get(STATUS)}: {key} -> {value}")
                    changed += 1
                    continue
                base = {"p": project["id"], "i": item["id"], "f": field}
                if value is None:
                    graphql(token, CLEAR, base)
                elif key == "effective":
                    graphql(token, SET, {**base, "v": {"number": value}})
                else:
                    graphql(token, SET, {**base, "v": {"text": value}})
                changed += 1
        page = project["items"]["pageInfo"]
        if not page["hasNextPage"]:
            break
        after = page["endCursor"]
    print(f"Campi aggiornati: {changed}")


if __name__ == "__main__":
    sys.exit(main())
