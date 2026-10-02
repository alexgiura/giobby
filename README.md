# Giobby API

Go API for the Giobby business platform, with PostgreSQL and a small API explorer.

## Components

- **backend** — HTTP API (Go, Gorilla Mux, JWT). Listens on port `8080`.
- **db** — PostgreSQL 16. Schema is created from `db/init_scripts` on the first start.
- **api-explorer** — Swagger UI for the API. Available on port `3000`.

## Run

Create `backend/.env` with the database and JWT settings, then start the stack:

```bash
docker compose up -d --build
```

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
