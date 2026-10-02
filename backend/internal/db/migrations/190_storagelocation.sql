-- StorageLocation

CREATE TABLE IF NOT EXISTS storage_locations (
    id_location     TEXT PRIMARY KEY,
    id_storage      TEXT NOT NULL REFERENCES storages(id),
    company_id      BIGINT NOT NULL DEFAULT 1,
    description     TEXT
);

CREATE INDEX IF NOT EXISTS idx_storage_locations_storage ON storage_locations (id_storage);

INSERT INTO storage_locations (id_location, id_storage, description) VALUES
    ('LOC01', 'MAG01', 'Scaffale A1')
ON CONFLICT (id_location) DO NOTHING;
