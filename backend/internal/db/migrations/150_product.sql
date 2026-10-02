-- Product (includes variants)

CREATE SEQUENCE IF NOT EXISTS product_code_seq START 1;

CREATE TABLE IF NOT EXISTS products (
    id                      TEXT PRIMARY KEY,
    company_id              BIGINT NOT NULL DEFAULT 1,
    barcode                 TEXT,
    basic_code              TEXT,
    id_type                 INTEGER DEFAULT 1,
    type_desc               TEXT DEFAULT 'Prodotto',
    description             TEXT NOT NULL,
    description_it          TEXT,
    description_en          TEXT,
    description_es          TEXT,
    description_bg          TEXT,
    note                    TEXT,
    id_material_group       INTEGER REFERENCES product_groups(id),
    um                      TEXT REFERENCES units_of_measure(um),
    sales_enabled           BOOLEAN NOT NULL DEFAULT true,
    locked                  BOOLEAN NOT NULL DEFAULT false,
    stock_enabled           BOOLEAN NOT NULL DEFAULT false,
    pricelist_enabled       BOOLEAN NOT NULL DEFAULT true,
    sales_price             DOUBLE PRECISION DEFAULT 0,
    sales_price_include_vat DOUBLE PRECISION DEFAULT 0,
    id_vat                  TEXT,
    purchase_price          DOUBLE PRECISION DEFAULT 0,
    manage_variant          BOOLEAN NOT NULL DEFAULT false,
    gross_weight            DOUBLE PRECISION DEFAULT 0,
    net_weight              DOUBLE PRECISION DEFAULT 0,
    dim_height              DOUBLE PRECISION DEFAULT 0,
    dim_length              DOUBLE PRECISION DEFAULT 0,
    dim_width               DOUBLE PRECISION DEFAULT 0,
    box_qty                 DOUBLE PRECISION DEFAULT 0,
    manufacturer            TEXT,
    default_storage         TEXT,
    repo_media_id           TEXT,
    create_date             TIMESTAMPTZ NOT NULL DEFAULT now(),
    modify_date             TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted                 BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_products_description ON products (description);
CREATE INDEX IF NOT EXISTS idx_products_group ON products (id_material_group);
CREATE INDEX IF NOT EXISTS idx_products_company ON products (company_id);

CREATE TABLE IF NOT EXISTS product_variants (
    id                      SERIAL PRIMARY KEY,
    product_id              TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    attribute_ids_concat    TEXT,
    barcode                 TEXT,
    delta_price             DOUBLE PRECISION DEFAULT 0,
    description             TEXT
);

CREATE INDEX IF NOT EXISTS idx_product_variants_product ON product_variants (product_id);

INSERT INTO products (
    id, description, description_it, um, sales_enabled, pricelist_enabled, sales_price, id_material_group, id_vat
) VALUES
    ('M00001', 'Prodotto demo', 'Prodotto demo', 'PZ', true, true, 10.00, 1, 'TVA19')
ON CONFLICT (id) DO NOTHING;
