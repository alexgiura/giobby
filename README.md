# Giobby API

Go API for the Giobby business platform, with PostgreSQL and a small API explorer.
Production runs on Coolify (VPN only): see [docs/DEPLOY-COOLIFY.md](docs/DEPLOY-COOLIFY.md).

## Components

- **backend** — HTTP API (Go, Gorilla Mux, JWT). Listens on port `8080`.
- **migrate** — applies the versioned migrations in `backend/internal/db/migrations` before the API starts, on every start.
- **db** — PostgreSQL 17 (local only; production uses the shared Postgres of cloud4job-infra).
- **api-explorer** — Swagger UI for the API. Available on port `3000`.

## Run

```bash
git config core.hooksPath .githooks
cp backend/.env.example backend/.env
docker compose up -d --build
```

An existing local volume from PostgreSQL 16 must be recreated once: `docker compose down -v`.

| Service      | URL                          |
|--------------|------------------------------|
| API          | http://localhost:8080        |
| Health check | http://localhost:8080/health |
| API explorer | http://localhost:3000        |
| PostgreSQL   | localhost:5432               |

The OpenAPI document is served at `http://localhost:8080/api/openapi.json`.

## API areas

Authentication, reference data, company settings, contacts, customers, vendors, products, warehouse, sales, purchases, accounting, CRM, calendar, users, messages, and notifications.

Routes are registered under `/api`.
