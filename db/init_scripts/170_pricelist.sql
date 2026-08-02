-- Pricelist

CREATE TABLE IF NOT EXISTS pricelist_schemes (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    description     TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS pricelists (
    id                  SERIAL PRIMARY KEY,
    company_id          BIGINT NOT NULL DEFAULT 1,
    id_pricelist_scheme INTEGER REFERENCES pricelist_schemes(id),
    description         TEXT NOT NULL,
    type                TEXT,
    note                TEXT,
    valid_since         TIMESTAMPTZ,
    valid_until         TIMESTAMPTZ,
    priority            INTEGER DEFAULT 0,
    source_id           TEXT,
    deleted             BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS pricelist_rows (
    id                  SERIAL PRIMARY KEY,
    pricelist_id        INTEGER NOT NULL REFERENCES pricelists(id) ON DELETE CASCADE,
    material_id         TEXT REFERENCES products(id),
    prod_description    TEXT,
    price               DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_pricelist_rows_list ON pricelist_rows (pricelist_id);

INSERT INTO pricelist_schemes (description) VALUES ('Listino base');

INSERT INTO pricelists (id_pricelist_scheme, description, type, priority)
SELECT 1, 'Listino vendita', '103', 1
WHERE NOT EXISTS (SELECT 1 FROM pricelists WHERE description = 'Listino vendita');

INSERT INTO pricelist_rows (pricelist_id, material_id, prod_description, price)
SELECT p.id, 'M00001', 'Prodotto demo', 10.00
FROM pricelists p
WHERE p.description = 'Listino vendita'
  AND NOT EXISTS (
      SELECT 1 FROM pricelist_rows r WHERE r.pricelist_id = p.id AND r.material_id = 'M00001'
  );
