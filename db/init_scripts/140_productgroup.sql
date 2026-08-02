-- ProductGroup

CREATE TABLE IF NOT EXISTS product_groups (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    id_parent       INTEGER REFERENCES product_groups(id),
    description_it  TEXT,
    description_en  TEXT,
    description_es  TEXT,
    description_bg  TEXT
);

INSERT INTO product_groups (id, description_it, description_en) VALUES
    (1, 'Generale', 'General'),
    (2, 'Servizi', 'Services')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('product_groups', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM product_groups), 1));
