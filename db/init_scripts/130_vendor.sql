-- Vendor

CREATE TABLE IF NOT EXISTS vendors (
    id                  TEXT PRIMARY KEY,
    contact_id          INTEGER NOT NULL UNIQUE REFERENCES contacts(id) ON DELETE CASCADE,
    id_payment_term     INTEGER,
    id_user_agent1      INTEGER,
    id_user_agent2      INTEGER,
    currency            TEXT,
    credit              DOUBLE PRECISION DEFAULT 0,
    dsc1                DOUBLE PRECISION DEFAULT 0,
    dsc2                DOUBLE PRECISION DEFAULT 0,
    dsc3                DOUBLE PRECISION DEFAULT 0,
    dsc4                DOUBLE PRECISION DEFAULT 0,
    price_list_id       TEXT,
    price_list_type     TEXT,
    price_list_enabled  BOOLEAN NOT NULL DEFAULT false,
    bank                TEXT,
    iban                TEXT,
    swift               TEXT,
    deleted             BOOLEAN NOT NULL DEFAULT false
);

INSERT INTO vendors (id, contact_id, currency)
SELECT 'V00001', id, 'RON' FROM contacts WHERE email = 'mario@example.com'
ON CONFLICT (id) DO NOTHING;

UPDATE contacts SET is_vendor = true WHERE email = 'mario@example.com';
