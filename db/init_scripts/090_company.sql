-- Company

CREATE TABLE IF NOT EXISTS companies (
    id              BIGINT PRIMARY KEY,
    name            TEXT NOT NULL,
    vat_number      TEXT,
    email           TEXT,
    phone           TEXT,
    address         TEXT,
    city            TEXT,
    zip             TEXT,
    country         TEXT REFERENCES countries(id),
    base_currency   TEXT REFERENCES currencies(code),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO companies (id, name, vat_number, email, city, country, base_currency) VALUES
    (1, 'Demo SRL', 'RO12345678', 'info@demo.local', 'București', 'RO', 'RON')
ON CONFLICT (id) DO NOTHING;
