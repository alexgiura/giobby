#!/bin/sh
# Checks docker-compose.prod.yml without deploying: required variables fail fast,
# no database or volume in the stack, and the only published port is VPN-only.
set -eu
cd "$(dirname "$0")/.."
FILE=docker-compose.prod.yml
REQUIRED="POSTGRES_DB_HOST POSTGRES_DB_NAME POSTGRES_DB_USER POSTGRES_DB_PASSWORD JWT_SECRET MIGRATION_DB_USER MIGRATION_DB_PASSWORD MIGRATION_APP_ROLE"

set_all() { for name in $REQUIRED; do export "$name=check-value"; done; }
fail() { echo "FAIL: $*" >&2; exit 1; }

set_all
docker compose -f "$FILE" config -q || fail "config with all variables"

for v in $REQUIRED; do
  set_all
  unset "$v"
  if err=$(docker compose -f "$FILE" config -q 2>&1); then
    fail "config without $v should fail"
  fi
  echo "$err" | grep -q "$v" || fail "error without $v does not name it: $err"
done

set_all
json=$(docker compose -f "$FILE" config --format json)
echo "$json" | jq -e '.services | has("db") | not' >/dev/null || fail "a db service is defined"
echo "$json" | jq -e '(.volumes // {}) | length == 0' >/dev/null || fail "volumes are defined"
ports=$(echo "$json" | jq -c '[.services[] | (.ports // [])[] | {host_ip, published, target}]')
[ "$ports" = '[{"host_ip":"10.8.0.1","published":"2001","target":80}]' ] || fail "published ports: $ports"

echo "prod compose OK"
