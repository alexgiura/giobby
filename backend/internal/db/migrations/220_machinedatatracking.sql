-- MachineDataTracking

CREATE TABLE IF NOT EXISTS machine_data_tracking (
    id                  SERIAL PRIMARY KEY,
    company_id          BIGINT NOT NULL DEFAULT 1,
    id_machine          TEXT,
    id_operator         TEXT,
    materials           TEXT,
    id_lot              TEXT,
    box_qty             DOUBLE PRECISION DEFAULT 0,
    box_number          DOUBLE PRECISION DEFAULT 0,
    production_waste    DOUBLE PRECISION DEFAULT 0,
    barcode             TEXT,
    start_time          TEXT,
    end_time            TEXT,
    duration            TEXT,
    machine_status      TEXT,
    alarm_code          TEXT,
    alarm_time          TEXT,
    track_date          TEXT,
    id_attachment       TEXT,
    deleted             BOOLEAN NOT NULL DEFAULT false
);
