-- PurchaseDocument (Task 10): purchase orders, goods receipt, vendor invoices, payments.

CREATE SEQUENCE IF NOT EXISTS purchase_doc_number_seq START 1000;

CREATE TABLE IF NOT EXISTS purchase_documents (
    id                  SERIAL PRIMARY KEY,
    company_id          BIGINT NOT NULL DEFAULT 1,
    doc_kind            TEXT NOT NULL,
    id_document_type    INTEGER,
    id_contact          INTEGER,
    id_customer         TEXT,
    id_vendor           TEXT,
    doc_date            TIMESTAMPTZ NOT NULL DEFAULT now(),
    doc_status          TEXT NOT NULL DEFAULT 'CREATED',
    doc_currency        TEXT DEFAULT 'EUR',
    tdsc1               DOUBLE PRECISION DEFAULT 0,
    tdsc2               DOUBLE PRECISION DEFAULT 0,
    tdsc3               DOUBLE PRECISION DEFAULT 0,
    tdsc4               DOUBLE PRECISION DEFAULT 0,
    id_bu               TEXT,
    id_numerator        INTEGER,
    doc_number          TEXT,
    sequential_number   INTEGER,
    note                TEXT,
    internal_note       TEXT,
    id_payment_term     INTEGER,
    reference           TEXT,
    dest_id_storage     TEXT,
    source_document_id  INTEGER REFERENCES purchase_documents(id),
    pdf_data            BYTEA,
    deleted             BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_purchase_documents_kind ON purchase_documents (doc_kind);
CREATE INDEX IF NOT EXISTS idx_purchase_documents_vendor ON purchase_documents (id_vendor);
CREATE INDEX IF NOT EXISTS idx_purchase_documents_contact ON purchase_documents (id_contact);
CREATE INDEX IF NOT EXISTS idx_purchase_documents_date ON purchase_documents (doc_date);

CREATE TABLE IF NOT EXISTS purchase_document_rows (
    id                          SERIAL PRIMARY KEY,
    document_id                 INTEGER NOT NULL REFERENCES purchase_documents(id) ON DELETE CASCADE,
    id_pos                      INTEGER NOT NULL,
    id_material                 TEXT,
    id_attribute_combination    INTEGER DEFAULT 0,
    id_pos_type                 INTEGER DEFAULT 1,
    description                 TEXT NOT NULL,
    price                       DOUBLE PRECISION NOT NULL DEFAULT 0,
    quantity                    DOUBLE PRECISION NOT NULL DEFAULT 1,
    packages                    INTEGER DEFAULT 1,
    dsc1                        DOUBLE PRECISION DEFAULT 0,
    dsc2                        DOUBLE PRECISION DEFAULT 0,
    dsc3                        DOUBLE PRECISION DEFAULT 0,
    dsc4                        DOUBLE PRECISION DEFAULT 0,
    id_vat                      TEXT,
    um                          TEXT
);

CREATE INDEX IF NOT EXISTS idx_purchase_document_rows_doc ON purchase_document_rows (document_id);

CREATE TABLE IF NOT EXISTS purchase_document_payments (
    id              SERIAL PRIMARY KEY,
    document_id     INTEGER NOT NULL REFERENCES purchase_documents(id) ON DELETE CASCADE,
    rate_number     INTEGER DEFAULT 1,
    description     TEXT,
    amount          DOUBLE PRECISION NOT NULL DEFAULT 0,
    paid_amount     DOUBLE PRECISION DEFAULT 0,
    due_date        TIMESTAMPTZ,
    payment_date    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_purchase_document_payments_doc ON purchase_document_payments (document_id);

CREATE TABLE IF NOT EXISTS purchase_document_attachments (
    id              SERIAL PRIMARY KEY,
    document_id     INTEGER NOT NULL REFERENCES purchase_documents(id) ON DELETE CASCADE,
    filename        TEXT NOT NULL,
    content_type    TEXT,
    data            BYTEA NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS vendor_payments (
    id                  SERIAL PRIMARY KEY,
    company_id          BIGINT NOT NULL DEFAULT 1,
    id_document_type    INTEGER,
    reference           TEXT,
    id_agent            INTEGER,
    total_paid_ic       DOUBLE PRECISION DEFAULT 0,
    total_invoice_ic    DOUBLE PRECISION DEFAULT 0,
    total_remaining_ic  DOUBLE PRECISION DEFAULT 0,
    doc_date            TIMESTAMPTZ NOT NULL DEFAULT now(),
    note                TEXT,
    doc_number          TEXT,
    deleted             BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Demo purchase order for vendor Mario
INSERT INTO purchase_documents (
    doc_kind, id_contact, id_vendor, doc_currency, id_bu, id_numerator, doc_number, sequential_number, doc_status
)
SELECT 'order', c.id, 'V00001', 'RON', 'U1', 1, 'PO-1000', 1000, 'CREATED'
FROM contacts c WHERE c.email = 'mario@example.com'
  AND NOT EXISTS (SELECT 1 FROM purchase_documents WHERE doc_kind = 'order' AND doc_number = 'PO-1000');

INSERT INTO purchase_document_rows (document_id, id_pos, id_material, description, price, quantity, id_vat, um)
SELECT d.id, 1, 'M00001', 'Materiale demo', 8.50, 10, 'TVA19', 'PZ'
FROM purchase_documents d
WHERE d.doc_number = 'PO-1000'
  AND NOT EXISTS (SELECT 1 FROM purchase_document_rows r WHERE r.document_id = d.id);

INSERT INTO purchase_document_payments (document_id, rate_number, description, amount, due_date)
SELECT d.id, 1, 'Rata unica', 85.00, now() + interval '30 days'
FROM purchase_documents d
WHERE d.doc_number = 'PO-1000'
  AND NOT EXISTS (SELECT 1 FROM purchase_document_payments p WHERE p.document_id = d.id);
