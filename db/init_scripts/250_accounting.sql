-- Accounting (Task 11): manual account movements.

CREATE TABLE IF NOT EXISTS account_movement_documents (
    id                      SERIAL PRIMARY KEY,
    company_id              BIGINT NOT NULL DEFAULT 1,
    id_bu                   TEXT,
    id_bup                  TEXT,
    id_numerator            INTEGER,
    id_accountmovement_type INTEGER,
    doc_date                TIMESTAMPTZ NOT NULL DEFAULT now(),
    reg_date                TIMESTAMPTZ NOT NULL DEFAULT now(),
    id_account_center       INTEGER REFERENCES account_centers(id),
    description             TEXT,
    entry_type              TEXT,
    deleted                 BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_account_movement_docs_company ON account_movement_documents (company_id);
CREATE INDEX IF NOT EXISTS idx_account_movement_docs_date ON account_movement_documents (doc_date);
CREATE INDEX IF NOT EXISTS idx_account_movement_docs_bu ON account_movement_documents (id_bu);

CREATE TABLE IF NOT EXISTS account_movement_rows (
    id                  SERIAL PRIMARY KEY,
    document_id         INTEGER NOT NULL REFERENCES account_movement_documents(id) ON DELETE CASCADE,
    line_no             INTEGER NOT NULL DEFAULT 1,
    description         TEXT,
    from_account_code   TEXT,
    to_account_code     TEXT,
    account_code        TEXT,
    account_sign        TEXT,
    amount              DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_account_movement_rows_doc ON account_movement_rows (document_id);
CREATE INDEX IF NOT EXISTS idx_account_movement_rows_code ON account_movement_rows (account_code);

-- Demo manual movement: cash to suppliers (E entry)
INSERT INTO account_movement_documents (
    id, company_id, id_bu, id_bup, id_numerator, id_accountmovement_type,
    doc_date, reg_date, id_account_center, description, entry_type
)
SELECT 1, 1, 'U1', 'P1', 1, 1, now(), now(), 1, 'Registrazione manuale demo', 'E'
WHERE NOT EXISTS (SELECT 1 FROM account_movement_documents WHERE id = 1);

INSERT INTO account_movement_rows (document_id, line_no, description, account_code, account_sign, amount)
SELECT 1, 1, 'Dare cassa', '1010', 'D', 100.00
WHERE NOT EXISTS (SELECT 1 FROM account_movement_rows WHERE document_id = 1 AND line_no = 1);

INSERT INTO account_movement_rows (document_id, line_no, description, account_code, account_sign, amount)
SELECT 1, 2, 'Avere fornitori', '2010', 'A', 100.00
WHERE NOT EXISTS (SELECT 1 FROM account_movement_rows WHERE document_id = 1 AND line_no = 2);

SELECT setval(
    pg_get_serial_sequence('account_movement_documents', 'id'),
    GREATEST((SELECT COALESCE(MAX(id), 1) FROM account_movement_documents), 1)
);
