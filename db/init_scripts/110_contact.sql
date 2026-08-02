-- Contact

CREATE TABLE IF NOT EXISTS contact_sources (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS contacts (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    type            TEXT NOT NULL DEFAULT 'COMPANY',
    name            TEXT NOT NULL,
    last_name       TEXT,
    address         TEXT,
    city            TEXT,
    country         TEXT,
    state           TEXT,
    postal_code     TEXT,
    pr              TEXT,
    fiscal_code     TEXT,
    vat_code        TEXT,
    phone1          TEXT,
    phone2          TEXT,
    mobile          TEXT,
    fax             TEXT,
    email           TEXT,
    email2          TEXT,
    visibility_type TEXT DEFAULT 'ALL',
    status          TEXT,
    id_user_owner   INTEGER,
    source          TEXT,
    note            TEXT,
    photo_data      BYTEA,
    photo_mime      TEXT,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    is_customer     BOOLEAN NOT NULL DEFAULT false,
    is_vendor       BOOLEAN NOT NULL DEFAULT false,
    is_lead         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_contacts_name ON contacts (name);
CREATE INDEX IF NOT EXISTS idx_contacts_email ON contacts (email);
CREATE INDEX IF NOT EXISTS idx_contacts_flags ON contacts (is_customer, is_vendor, deleted);

CREATE TABLE IF NOT EXISTS contact_offices (
    id              SERIAL PRIMARY KEY,
    contact_id      INTEGER NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    id_office_type  INTEGER,
    name            TEXT,
    address         TEXT,
    city            TEXT,
    pr              TEXT,
    postal_code     TEXT,
    country         TEXT,
    state           TEXT,
    email           TEXT,
    phone1          TEXT,
    phone2          TEXT,
    fax             TEXT,
    default_dest    BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS contact_subcontact_assocs (
    id              SERIAL PRIMARY KEY,
    parent_contact_id INTEGER NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    child_contact_id  INTEGER NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    id_contact_role   INTEGER,
    UNIQUE (parent_contact_id, child_contact_id)
);

INSERT INTO contact_sources (name) VALUES
    ('Web'),
    ('Fiera'),
    ('Referral')
ON CONFLICT (name) DO NOTHING;

INSERT INTO contacts (type, name, last_name, email, city, country, visibility_type, status, source) VALUES
    ('COMPANY', 'Acme SRL', NULL, 'info@acme.local', 'București', 'RO', 'ALL', 'ACTIVE', 'Web'),
    ('PRIVATE_PERSON', 'Mario', 'Rossi', 'mario@example.com', 'Roma', 'IT', 'ALL', 'ACTIVE', 'Fiera');

INSERT INTO contact_offices (contact_id, id_office_type, name, city, country, default_dest)
SELECT c.id, 1, 'Sede principale', c.city, c.country, true
FROM contacts c WHERE c.email = 'info@acme.local'
  AND NOT EXISTS (SELECT 1 FROM contact_offices o WHERE o.contact_id = c.id);
