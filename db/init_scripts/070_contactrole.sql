-- ContactRole

CREATE TABLE IF NOT EXISTS contact_roles (
    id              SERIAL PRIMARY KEY,
    description     TEXT NOT NULL
);

INSERT INTO contact_roles (description) VALUES
    ('Amministratore'),
    ('Commerciale'),
    ('Tecnico');
