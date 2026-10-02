-- Calendar (Task 13): calendars and events.

CREATE TABLE IF NOT EXISTS calendars (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    name            TEXT NOT NULL,
    color           TEXT DEFAULT '#3366FF',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS calendar_events (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    id_calendar     INTEGER NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    event           TEXT NOT NULL,
    place           TEXT,
    start_date      TIMESTAMPTZ NOT NULL,
    end_date        TIMESTAMPTZ NOT NULL,
    color           TEXT NOT NULL DEFAULT '#3366FF',
    privacy         TEXT NOT NULL DEFAULT 'PUBLIC',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_calendar_events_calendar ON calendar_events (id_calendar);
CREATE INDEX IF NOT EXISTS idx_calendar_events_start ON calendar_events (start_date);

INSERT INTO calendars (id, name, color)
SELECT 1, 'Calendario principale', '#3366FF'
WHERE NOT EXISTS (SELECT 1 FROM calendars WHERE id = 1);

INSERT INTO calendar_events (id_calendar, event, place, start_date, end_date, color, privacy)
SELECT 1, 'Riunione team', 'Sede centrale', now() + interval '1 day', now() + interval '1 day 1 hour', '#FF6600', 'PUBLIC'
WHERE NOT EXISTS (SELECT 1 FROM calendar_events WHERE event = 'Riunione team');

SELECT setval(pg_get_serial_sequence('calendars', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM calendars), 1));
SELECT setval(pg_get_serial_sequence('calendar_events', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM calendar_events), 1));
