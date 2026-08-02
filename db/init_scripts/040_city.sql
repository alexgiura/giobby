-- City

CREATE TABLE IF NOT EXISTS cities (
    id              SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    pr              TEXT,
    zip             TEXT,
    country         TEXT REFERENCES countries(id),
    state           TEXT
);

CREATE INDEX IF NOT EXISTS idx_cities_name ON cities (name);
CREATE INDEX IF NOT EXISTS idx_cities_country ON cities (country);

INSERT INTO cities (name, pr, zip, country, state) VALUES
    ('Roma', 'RM', '00100', 'IT', 'Lazio'),
    ('Milano', 'MI', '20100', 'IT', 'Lombardia'),
    ('București', 'B', '010001', 'RO', 'București');
