-- Currency

CREATE TABLE IF NOT EXISTS currencies (
    code            TEXT PRIMARY KEY,
    description     TEXT NOT NULL
);

INSERT INTO currencies (code, description) VALUES
    ('EUR', 'Euro'),
    ('RON', 'Leu românesc')
ON CONFLICT (code) DO NOTHING;
