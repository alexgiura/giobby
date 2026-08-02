-- Stock

CREATE TABLE IF NOT EXISTS stocks (
    id                          SERIAL PRIMARY KEY,
    id_storage                  TEXT REFERENCES storages(id),
    id_location                 TEXT REFERENCES storage_locations(id_location),
    id_lot                      TEXT,
    id_material                 TEXT REFERENCES products(id),
    um                          TEXT,
    quantity                    DOUBLE PRECISION NOT NULL DEFAULT 0,
    value                       DOUBLE PRECISION DEFAULT 0,
    number_of_containers        INTEGER DEFAULT 0,
    stock_state                 INTEGER DEFAULT 1,
    id_document_type            INTEGER DEFAULT 109,
    id_attribute_combination    INTEGER DEFAULT 0,
    attribute_combination_desc  TEXT,
    storage_desc                TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_stocks_material ON stocks (id_material);
CREATE INDEX IF NOT EXISTS idx_stocks_storage ON stocks (id_storage);

INSERT INTO stocks (
    id_storage, id_location, id_lot, id_material, um, quantity, value, storage_desc
)
SELECT 'MAG01', 'LOC01', 'L00001', 'M00001', 'PZ', 100, 10.00, 'Magazzino principale'
WHERE NOT EXISTS (SELECT 1 FROM stocks WHERE id_material = 'M00001' AND id_storage = 'MAG01');
