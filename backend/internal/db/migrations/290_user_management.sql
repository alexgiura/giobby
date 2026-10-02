-- User management (Task 15): Giobby users, auth profiles, agent commissions.
-- Distinct from auth `users` (010_user.sql) used for JWT login.

CREATE TABLE IF NOT EXISTS company_users (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    username        TEXT NOT NULL,
    password_hash   TEXT,
    email           TEXT,
    firstname       TEXT,
    lastname        TEXT,
    mobilephone     TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT false,
    is_agent1       BOOLEAN NOT NULL DEFAULT false,
    is_agent2       BOOLEAN NOT NULL DEFAULT false,
    is_employee     BOOLEAN NOT NULL DEFAULT false,
    image_data      BYTEA,
    image_mime      TEXT,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT company_users_username_unique UNIQUE (company_id, username)
);

CREATE INDEX IF NOT EXISTS idx_company_users_agents ON company_users (is_agent1, is_agent2);

CREATE TABLE IF NOT EXISTS user_auth_profiles (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER NOT NULL REFERENCES company_users(id) ON DELETE CASCADE,
    profile_name    TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, profile_name)
);

CREATE TABLE IF NOT EXISTS user_agent_commissions (
    id                  SERIAL PRIMARY KEY,
    company_id          BIGINT NOT NULL DEFAULT 1,
    user_id             INTEGER NOT NULL REFERENCES company_users(id) ON DELETE CASCADE,
    id_bu               TEXT,
    id_customer         TEXT,
    id_material_group   INTEGER,
    id_material         TEXT,
    commission_rate     DOUBLE PRECISION NOT NULL DEFAULT 0,
    description         TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_agent_commissions_user ON user_agent_commissions (user_id);

INSERT INTO company_users (id, username, email, firstname, lastname, is_active, is_agent1, is_employee)
SELECT 1, 'agent.demo', 'agent@example.com', 'Agent', 'Demo', true, true, true
WHERE NOT EXISTS (SELECT 1 FROM company_users WHERE id = 1);

INSERT INTO user_auth_profiles (user_id, profile_name)
SELECT 1, 'SALES'
WHERE NOT EXISTS (SELECT 1 FROM user_auth_profiles WHERE user_id = 1 AND profile_name = 'SALES');

INSERT INTO user_agent_commissions (user_id, id_bu, id_customer, id_material, commission_rate, description)
SELECT 1, 'U1', 'C00001', 'M00001', 5.0, 'Commissione demo'
WHERE NOT EXISTS (SELECT 1 FROM user_agent_commissions WHERE user_id = 1 AND id_material = 'M00001');

SELECT setval(pg_get_serial_sequence('company_users', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM company_users), 1));
