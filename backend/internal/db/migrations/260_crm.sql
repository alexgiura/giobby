-- Crm (Task 12): CRM accounts, tasks, activities, attachments.

CREATE TABLE IF NOT EXISTS crm_accounts (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    contact_id      INTEGER NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (company_id, contact_id)
);

CREATE INDEX IF NOT EXISTS idx_crm_accounts_contact ON crm_accounts (contact_id);

CREATE TABLE IF NOT EXISTS crm_tasks (
    id                              SERIAL PRIMARY KEY,
    company_id                      BIGINT NOT NULL DEFAULT 1,
    id_root                         INTEGER NOT NULL,
    id_parent                       INTEGER NOT NULL,
    type                            TEXT NOT NULL DEFAULT 'TASK_CRM',
    name                            TEXT,
    description                     TEXT,
    place                           TEXT,
    status                          TEXT DEFAULT 'OPEN',
    start_date                      TIMESTAMPTZ,
    end_date                        TIMESTAMPTZ,
    create_date                     TIMESTAMPTZ NOT NULL DEFAULT now(),
    id_doc_type                     INTEGER,
    id_doc                          TEXT,
    doc_description                 TEXT,
    priority                        INTEGER DEFAULT 0,
    privacy                         TEXT,
    id_user_assignee                INTEGER,
    user_assignee_description       TEXT,
    id_contact_representative       INTEGER,
    contact_representative_desc     TEXT,
    quantity                        DOUBLE PRECISION DEFAULT 0,
    amount                          DOUBLE PRECISION DEFAULT 0,
    probability                     DOUBLE PRECISION DEFAULT 0,
    forecast_revenue                DOUBLE PRECISION DEFAULT 0,
    deleted                         BOOLEAN NOT NULL DEFAULT false,
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_crm_tasks_root ON crm_tasks (id_root);
CREATE INDEX IF NOT EXISTS idx_crm_tasks_parent ON crm_tasks (id_parent);
CREATE INDEX IF NOT EXISTS idx_crm_tasks_type ON crm_tasks (type);

CREATE TABLE IF NOT EXISTS crm_activities (
    id                      SERIAL PRIMARY KEY,
    company_id              BIGINT NOT NULL DEFAULT 1,
    id_root                 INTEGER NOT NULL,
    id_parent               INTEGER NOT NULL,
    id_user                 INTEGER,
    id_attachment           TEXT,
    attachment_file_name    TEXT,
    name                    TEXT,
    description             TEXT,
    type                    TEXT,
    id_document_type        INTEGER,
    id_document_type_ext    TEXT,
    id_doc                  TEXT,
    doc_description         TEXT,
    deleted                 BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_crm_activities_parent ON crm_activities (id_parent);

CREATE TABLE IF NOT EXISTS crm_attachments (
    id              SERIAL PRIMARY KEY,
    id_parent       INTEGER NOT NULL,
    filename        TEXT NOT NULL,
    content_type    TEXT,
    data            BYTEA NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- CRM on Acme contact
INSERT INTO crm_accounts (contact_id, name)
SELECT c.id, 'Acme CRM'
FROM contacts c WHERE c.email = 'info@acme.local'
  AND NOT EXISTS (SELECT 1 FROM crm_accounts a WHERE a.contact_id = c.id);

INSERT INTO crm_tasks (
    id, company_id, id_root, id_parent, type, name, description, status, amount, probability
)
SELECT 1, 1, a.id, a.id, 'TASK_CRM_OPPORTUNITY', 'Opportunità demo', 'Prima opportunità', 'OPEN', 5000, 60
FROM crm_accounts a
JOIN contacts c ON c.id = a.contact_id AND c.email = 'info@acme.local'
WHERE NOT EXISTS (SELECT 1 FROM crm_tasks WHERE id = 1);

INSERT INTO crm_activities (id_root, id_parent, type, name, description)
SELECT a.id, a.id, 'NOTE', 'Nota iniziale', 'Contatto CRM attivato'
FROM crm_accounts a
JOIN contacts c ON c.id = a.contact_id AND c.email = 'info@acme.local'
WHERE NOT EXISTS (
    SELECT 1 FROM crm_activities act WHERE act.id_parent = a.id AND act.name = 'Nota iniziale'
);

SELECT setval(pg_get_serial_sequence('crm_accounts', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM crm_accounts), 1));
SELECT setval(pg_get_serial_sequence('crm_tasks', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM crm_tasks), 1));
