-- OfficeType

CREATE TABLE IF NOT EXISTS office_types (
    id              SERIAL PRIMARY KEY,
    description     TEXT NOT NULL
);

INSERT INTO office_types (description) VALUES
    ('Sede legale'),
    ('Sede operativa'),
    ('Magazzino');
