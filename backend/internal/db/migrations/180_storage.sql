-- Storage

CREATE TABLE IF NOT EXISTS storages (
    id              TEXT PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    description     TEXT NOT NULL,
    managed         INTEGER NOT NULL DEFAULT 0,
    deleted         BOOLEAN NOT NULL DEFAULT false
);

INSERT INTO storages (id, description, managed) VALUES
    ('MAG01', 'Magazzino principale', 1)
ON CONFLICT (id) DO NOTHING;
