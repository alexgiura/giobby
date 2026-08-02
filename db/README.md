# Database — Giobby API (PostgreSQL)

Schema inițială prin **init scripts** (la primul start Postgres cu volum gol). Fără `golang-migrate`.

## Structură

```
db/
├── Dockerfile
└── init_scripts/
    ├── README.md              # mapare swagger tag → fișier SQL
    ├── 000_extensions.sql
    ├── 010_user.sql
    ├── 020_country.sql
    ├── …                      # un fișier per domeniu swagger
    └── 220_machinedatatracking.sql
```

Vezi [init_scripts/README.md](init_scripts/README.md) pentru lista completă swagger ↔ SQL.

## Docker (backend + Postgres)

```bash
docker compose up -d --build
```

- **db** — Postgres 16, rulează `init_scripts/*.sql` o singură dată (ordine alfabetică)
- **backend** — Go API pe `:8080`, se conectează la `db:5432`

### Reset schema (după modificări init_scripts)

```bash
docker compose down -v
docker compose up -d --build
```

## Local fără Docker

Rulează toate fișierele în ordine:

```bash
for f in db/init_scripts/*.sql; do
  psql "postgres://postgres:postgres@localhost:5432/giobby_bff?sslmode=disable" -f "$f"
done
```

## Evoluție schema

Adaugă un fișier nou per domeniu swagger (`230_saledocument.sql`, …) sau reintrodu migrări când ai nevoie de upgrade fără reset volum.

Pentru development cu Docker, cel mai simplu: `docker compose down -v` după schimbări majore de schema.
