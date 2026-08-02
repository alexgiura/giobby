-- Um (units of measure)

CREATE TABLE IF NOT EXISTS units_of_measure (
    um                      TEXT PRIMARY KEY,
    description_it          TEXT,
    description_en          TEXT,
    description_es          TEXT,
    description_bg          TEXT,
    short_description_it    TEXT,
    short_description_en    TEXT,
    short_description_es    TEXT,
    short_description_bg    TEXT,
    deleted_at              TIMESTAMPTZ
);

INSERT INTO units_of_measure (um, description_it, short_description_it) VALUES
    ('PZ', 'Pezzi', 'PZ'),
    ('KG', 'Chilogrammi', 'KG'),
    ('M2', 'Metri quadri', 'M2')
ON CONFLICT (um) DO NOTHING;
