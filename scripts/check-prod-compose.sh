#!/usr/bin/env bash
# Checks the production compose without deploying it (Cloud4Job kit).
#   scripts/check-prod-compose.sh [--deploy public|vpn] [FILE]   (default: docker-compose.prod.yml)
# Run from the project root; DEPLOY defaults to .cloud4job/project.env. Needs docker compose and jq.
set -euo pipefail

PROJECT_NAME="giobby"
VPN_IP="10.8.0.1"
GENERIC_NAMES="backend api app web frontend server db database redis"

deploy="" file=docker-compose.prod.yml
while [ $# -gt 0 ]; do
  case "$1" in
    --deploy) deploy=$2; shift 2 ;;
    *) file=$1; shift ;;
  esac
done
fail() { echo "FAIL: $*" >&2; exit 1; }

if [ -z "$deploy" ] && [ -f .cloud4job/project.env ]; then
  deploy=$(grep -E '^DEPLOY=' .cloud4job/project.env | cut -d= -f2 || true)
fi
[ -n "$deploy" ] || fail "tipo di deploy sconosciuto: usare --deploy public|vpn|none o DEPLOY in .cloud4job/project.env"
[ -f "$file" ] || fail "$file non trovato"

# Coolify stores the text after ${VAR:? as the variable's value: only the bare form is allowed.
if grep -nE '^[^#]*\$\{[A-Z_][A-Z0-9_]*:\?[^}]' "$file"; then
  fail "variabili obbligatorie: usare \${VAR:?} senza messaggio (Coolify salva il messaggio come valore)"
fi

required=$(grep -vE '^[[:space:]]*#' "$file" | grep -oE '\$\{[A-Z_][A-Z0-9_]*:\?\}' | sed -E 's/\$\{([A-Z_][A-Z0-9_]*):\?\}/\1/' | sort -u || true)
set_all() { local n; for n in $required; do export "$n=check-value"; done; }
# A local .env would fill required variables and hide the check: never read it.
compose() { docker compose --env-file /dev/null -f "$file" "$@"; }

set_all
compose config -q || fail "configurazione non valida"
for v in $required; do
  set_all; unset "$v"
  if err=$(compose config -q 2>&1); then fail "senza $v la configurazione dovrebbe fallire"; fi
  echo "$err" | grep -q "$v" || fail "l'errore senza $v non nomina la variabile: $err"
done
[ -z "$required" ] || echo "variabili obbligatorie: $(echo "$required" | tr '\n' ' ')"

set_all
json=$(compose config --format json)

echo "$json" | jq -e '[.services[] | (.image // "") | test("(^|/)(postgres|postgresql|postgis|mysql|mariadb|mongo|mongodb)(:|@|$)")] | any | not' >/dev/null \
  || fail "servizio database nel compose: usare il Postgres condiviso"
echo "$json" | jq -e '((.volumes // {}) | length == 0) and ([.services[] | (.volumes // [])[]] | length == 0)' >/dev/null \
  || fail "volumi non ammessi nel compose di produzione: i dati stanno nel database condiviso"

ports=$(echo "$json" | jq -c '[.services[] | (.ports // [])[] | {host_ip, published}]')
case "$deploy" in
  public) [ "$ports" = "[]" ] || fail "porte pubblicate non ammesse con deploy public (il dominio passa da Traefik): $ports" ;;
  vpn)
    [ "$ports" != "[]" ] || fail "deploy vpn senza porta: serve \"$VPN_IP:<porta>:<porta interna>\""
    echo "$ports" | jq -e --arg ip "$VPN_IP" 'all(.host_ip == $ip)' >/dev/null \
      || fail "porta non legata a $VPN_IP (mai 0.0.0.0): $ports" ;;
  none) ;;
  *) fail "deploy sconosciuto: $deploy" ;;
esac

for svc in $(echo "$json" | jq -r '.services | keys[]'); do
  case "$svc" in "$PROJECT_NAME"-*) ;; *) fail "servizio $svc: il nome deve iniziare con $PROJECT_NAME- (rete condivisa di Coolify)" ;; esac
done

# A generic name as a host, alone or inside a URL (http://backend:8080, user@db:5432).
for name in $GENERIC_NAMES; do
  hit=$(echo "$json" | jq -r --arg n "$name" '.services[] | (.environment // {}) | to_entries[]
    | select((.value // "") | test("(^|//|@)" + $n + "(:|/|$)")) | .key')
  [ -z "$hit" ] || fail "la variabile $hit punta al nome generico '$name': usare $PROJECT_NAME-<servizio>"
done

for svc in $(echo "$json" | jq -r '.services | to_entries[] | select(.value.restart != "no") | .key'); do
  echo "$json" | jq -e --arg s "$svc" '.services[$s].healthcheck.test' >/dev/null \
    || fail "il servizio $svc non ha un healthcheck (senza, Coolify mostra lo stato unknown)"
done

echo "prod compose OK"
