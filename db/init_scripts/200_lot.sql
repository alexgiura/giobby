-- Lot

CREATE TABLE IF NOT EXISTS lots (
    id_lot          TEXT PRIMARY KEY,
    id_material     TEXT REFERENCES products(id),
    incoming_date   TIMESTAMPTZ,
    expire_date     TIMESTAMPTZ,
    reference       TEXT
);

INSERT INTO lots (id_lot, id_material, reference) VALUES
    ('L00001', 'M00001', 'Lot demo')
ON CONFLICT (id_lot) DO NOTHING;
