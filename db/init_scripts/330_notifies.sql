-- Notifies (ecommerce / social / tasks), emails, global notifications, social posts.

CREATE TABLE IF NOT EXISTS app_notifies (
    id              BIGSERIAL PRIMARY KEY,
    channel         TEXT NOT NULL, -- ecommerce | social | task
    title           TEXT NOT NULL DEFAULT '',
    body            TEXT NOT NULL DEFAULT '',
    unread          BOOLEAN NOT NULL DEFAULT true,
    payload         JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_app_notifies_channel ON app_notifies(channel);
CREATE INDEX IF NOT EXISTS idx_app_notifies_unread ON app_notifies(channel, unread);

CREATE TABLE IF NOT EXISTS social_posts (
    id              BIGSERIAL PRIMARY KEY,
    id_lang         TEXT DEFAULT 'en',
    title           TEXT NOT NULL DEFAULT '',
    body            TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS emails (
    id              BIGSERIAL PRIMARY KEY,
    hashcode        TEXT,
    subject         TEXT NOT NULL DEFAULT '',
    body            TEXT NOT NULL DEFAULT '',
    sender          TEXT,
    recipient       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS global_notifications (
    id              BIGSERIAL PRIMARY KEY,
    title           TEXT NOT NULL DEFAULT '',
    body            TEXT NOT NULL DEFAULT '',
    source_company  TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app_notifies (channel, title, body, unread)
SELECT 'ecommerce', 'Order shipped', 'Demo ecommerce notify', true
WHERE NOT EXISTS (SELECT 1 FROM app_notifies WHERE channel = 'ecommerce');

INSERT INTO app_notifies (channel, title, body, unread)
SELECT 'social', 'New mention', 'Demo social notify', true
WHERE NOT EXISTS (SELECT 1 FROM app_notifies WHERE channel = 'social');

INSERT INTO app_notifies (channel, title, body, unread)
SELECT 'task', 'Task updated', 'Demo task notify', true
WHERE NOT EXISTS (SELECT 1 FROM app_notifies WHERE channel = 'task');

INSERT INTO social_posts (id_lang, title, body)
SELECT 'en', 'Welcome', 'Public social post demo'
WHERE NOT EXISTS (SELECT 1 FROM social_posts LIMIT 1);

INSERT INTO emails (hashcode, subject, body, sender, recipient)
SELECT 'demo-hash', 'Welcome email', 'Demo email body', 'noreply@giobby.local', 'demo@example.com'
WHERE NOT EXISTS (SELECT 1 FROM emails LIMIT 1);

INSERT INTO global_notifications (title, body, source_company)
SELECT 'System notice', 'Global notification demo', 'External Co'
WHERE NOT EXISTS (SELECT 1 FROM global_notifications LIMIT 1);
