-- Messaging — Message + MessageGroups (Task 14).

CREATE TABLE IF NOT EXISTS message_groups (
    id              BIGSERIAL PRIMARY KEY,
    title           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS messages (
    id              BIGSERIAL PRIMARY KEY,
    id_group        BIGINT REFERENCES message_groups(id) ON DELETE CASCADE,
    body            TEXT NOT NULL DEFAULT '',
    unread          BOOLEAN NOT NULL DEFAULT true,
    user_from       TEXT,
    user_from_full_name TEXT,
    user_to_full_name   TEXT,
    id_transaction  INTEGER,
    id_channel      TEXT,
    company_to      INTEGER,
    sent_msg        BOOLEAN NOT NULL DEFAULT true,
    sent_date       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_group ON messages(id_group);
CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages(unread);

CREATE TABLE IF NOT EXISTS message_contacts (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    email           TEXT,
    internal_id     TEXT
);

INSERT INTO message_groups (id, title)
SELECT 1, 'Demo chat'
WHERE NOT EXISTS (SELECT 1 FROM message_groups WHERE id = 1);

INSERT INTO messages (id_group, body, unread, user_from, user_from_full_name, user_to_full_name, sent_msg)
SELECT 1, 'Hello from demo', true, 'demo', 'Demo User', 'Agent Demo', true
WHERE NOT EXISTS (SELECT 1 FROM messages LIMIT 1);

INSERT INTO message_contacts (name, email, internal_id)
SELECT 'Agent Demo', 'agent@example.com', '1'
WHERE NOT EXISTS (SELECT 1 FROM message_contacts LIMIT 1);

SELECT setval(pg_get_serial_sequence('message_groups', 'id'), (SELECT COALESCE(MAX(id), 1) FROM message_groups));
SELECT setval(pg_get_serial_sequence('messages', 'id'), (SELECT COALESCE(MAX(id), 1) FROM messages));
