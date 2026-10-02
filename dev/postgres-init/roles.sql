-- Local development only: synthetic passwords, matching backend/.env.example.
-- Mirrors production: giobby owns the database and runs migrations,
-- giobby_app is the API role (data only, granted by cmd/migrate).
CREATE ROLE giobby LOGIN PASSWORD 'dev-migration-password';
CREATE ROLE giobby_app LOGIN PASSWORD 'dev-app-password';
CREATE DATABASE giobby OWNER giobby;
REVOKE CONNECT ON DATABASE giobby FROM PUBLIC;
GRANT CONNECT ON DATABASE giobby TO giobby_app;
