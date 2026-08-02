-- CompanySettings

CREATE TABLE IF NOT EXISTS account_centers (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    type            TEXT,
    description     TEXT
);

CREATE INDEX IF NOT EXISTS idx_account_centers_company ON account_centers (company_id);

CREATE TABLE IF NOT EXISTS account_codes (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code            TEXT NOT NULL,
    description     TEXT,
    entry           TEXT,
    entry_type      TEXT,
    disabled        BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT account_codes_company_code_unique UNIQUE (company_id, code)
);

CREATE TABLE IF NOT EXISTS banks (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    description     TEXT NOT NULL,
    is_cashdesk     BOOLEAN NOT NULL DEFAULT false,
    account_code    TEXT,
    iban            TEXT,
    sia             TEXT,
    swift           TEXT,
    cuc             TEXT,
    deleted         BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS currency_rates (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    source_currency TEXT NOT NULL REFERENCES currencies(code),
    target_currency TEXT NOT NULL REFERENCES currencies(code),
    rate            DOUBLE PRECISION NOT NULL,
    effective_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    CONSTRAINT currency_rates_unique UNIQUE (company_id, source_currency, effective_date)
);

CREATE TABLE IF NOT EXISTS vat_rates (
    id              TEXT PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    id_bu           TEXT,
    country         TEXT,
    state           TEXT,
    contact_type    TEXT,
    tax_super_type  INTEGER,
    rate            DOUBLE PRECISION NOT NULL,
    description     TEXT
);

CREATE INDEX IF NOT EXISTS idx_vat_rates_company ON vat_rates (company_id);

CREATE TABLE IF NOT EXISTS business_units (
    id              TEXT PRIMARY KEY,
    company_id      BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name            TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS business_unit_points (
    id              TEXT PRIMARY KEY,
    bu_id           TEXT NOT NULL REFERENCES business_units(id) ON DELETE CASCADE,
    name            TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS document_numerators (
    id              SERIAL PRIMARY KEY,
    bup_id          TEXT NOT NULL REFERENCES business_unit_points(id) ON DELETE CASCADE,
    id_document_type INTEGER,
    id_plugin       TEXT,
    numerator       INTEGER NOT NULL DEFAULT 1
);

INSERT INTO account_centers (company_id, name, type, description) VALUES
    (1, 'Centro principale', 'COST', 'Centru de cost principal'),
    (1, 'Amministrazione', 'REVENUE', 'Centru venituri admin');

INSERT INTO account_codes (company_id, code, description, entry, entry_type) VALUES
    (1, '1010', 'Cassa', 'P', 'D'),
    (1, '2010', 'Fornitori', 'P', 'A'),
    (1, '4010', 'Ricavi vendite', 'E', 'R'),
    (1, '5010', 'Acquisti merci', 'E', 'C')
ON CONFLICT (company_id, code) DO NOTHING;

INSERT INTO banks (company_id, description, is_cashdesk, account_code, iban) VALUES
    (1, 'Banca Principale', false, '1010', 'RO49AAAA1B31007593840000'),
    (1, 'Cassa contanti', true, '1010', NULL);

INSERT INTO currency_rates (company_id, source_currency, target_currency, rate) VALUES
    (1, 'EUR', 'RON', 4.97)
ON CONFLICT DO NOTHING;

INSERT INTO vat_rates (id, company_id, id_bu, country, contact_type, tax_super_type, rate, description) VALUES
    ('IVA22', 1, 'U1', 'IT', 'C', 1, 22.0, 'IVA 22%'),
    ('IVA10', 1, 'U1', 'IT', 'C', 1, 10.0, 'IVA 10%'),
    ('TVA19', 1, 'U1', 'RO', 'C', 1, 19.0, 'TVA 19%')
ON CONFLICT (id) DO NOTHING;

INSERT INTO business_units (id, company_id, name) VALUES
    ('U1', 1, 'Unità principale')
ON CONFLICT (id) DO NOTHING;

INSERT INTO business_unit_points (id, bu_id, name) VALUES
    ('P1', 'U1', 'Sede centrale')
ON CONFLICT (id) DO NOTHING;

INSERT INTO document_numerators (bup_id, id_document_type, numerator) VALUES
    ('P1', 1, 100),
    ('P1', 2, 50);
