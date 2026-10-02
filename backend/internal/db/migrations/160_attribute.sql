-- Attribute

CREATE TABLE IF NOT EXISTS product_characteristics (
    id                  SERIAL PRIMARY KEY,
    company_id          BIGINT NOT NULL DEFAULT 1,
    attribute_name_it   TEXT,
    attribute_name_en   TEXT,
    attribute_name_es   TEXT,
    attribute_name_bg   TEXT
);

CREATE TABLE IF NOT EXISTS document_attributes (
    id                      SERIAL PRIMARY KEY,
    doc_type                INTEGER,
    id_doc                  TEXT NOT NULL,
    id_attribute            INTEGER,
    sub_id                  INTEGER DEFAULT 0,
    pos                     INTEGER DEFAULT 0,
    attribute_value         TEXT,
    attribute_name          TEXT,
    doc_type_attribute_value INTEGER,
    id_lang                 TEXT DEFAULT 'IT'
);

CREATE INDEX IF NOT EXISTS idx_document_attributes_doc ON document_attributes (id_doc);

INSERT INTO product_characteristics (attribute_name_it, attribute_name_en)
SELECT 'Taglia', 'Size'
WHERE NOT EXISTS (SELECT 1 FROM product_characteristics WHERE attribute_name_it = 'Taglia');

INSERT INTO product_characteristics (attribute_name_it, attribute_name_en)
SELECT 'Colore', 'Color'
WHERE NOT EXISTS (SELECT 1 FROM product_characteristics WHERE attribute_name_it = 'Colore');
