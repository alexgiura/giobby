-- Integration stubs audit (plugins / tilby). Optional persistence of inbound webhooks.

CREATE TABLE IF NOT EXISTS integration_events (
    id              BIGSERIAL PRIMARY KEY,
    provider        TEXT NOT NULL, -- bindcommerce | tilby
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
