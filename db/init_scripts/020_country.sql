-- Country

CREATE TABLE IF NOT EXISTS countries (
    id              TEXT PRIMARY KEY,
    description     TEXT NOT NULL
);

INSERT INTO countries (id, description) VALUES
    ('IT', 'Italia'),
    ('RO', 'Romania')
ON CONFLICT (id) DO NOTHING;
