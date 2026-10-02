-- Settings (Task 15): key-value application settings.

CREATE TABLE IF NOT EXISTS app_settings (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    key1            TEXT NOT NULL DEFAULT '',
    key2            TEXT NOT NULL DEFAULT '',
    key3            TEXT NOT NULL DEFAULT '',
    key4            TEXT NOT NULL DEFAULT '',
    key5            TEXT NOT NULL DEFAULT '',
    value           TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (company_id, key1, key2, key3, key4, key5)
);

INSERT INTO app_settings (key1, key2, value)
SELECT 'APP', 'LANGUAGE', 'it'
WHERE NOT EXISTS (SELECT 1 FROM app_settings WHERE key1 = 'APP' AND key2 = 'LANGUAGE');

INSERT INTO app_settings (key1, key2, value)
SELECT 'APP', 'TIMEZONE', 'Europe/Rome'
WHERE NOT EXISTS (SELECT 1 FROM app_settings WHERE key1 = 'APP' AND key2 = 'TIMEZONE');
