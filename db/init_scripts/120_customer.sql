-- Customer

CREATE TABLE IF NOT EXISTS customers (
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
    locked              BOOLEAN NOT NULL DEFAULT false,
    recipient_code      TEXT,
    deleted             BOOLEAN NOT NULL DEFAULT false
);

INSERT INTO customers (id, contact_id, currency, credit)
SELECT 'C00001', id, 'EUR', 1000 FROM contacts WHERE email = 'info@acme.local'
ON CONFLICT (id) DO NOTHING;

UPDATE contacts SET is_customer = true WHERE email = 'info@acme.local';
