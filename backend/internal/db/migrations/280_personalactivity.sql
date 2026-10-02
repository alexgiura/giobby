-- PersonalActivity (Task 13): personal to-do activities.

CREATE TABLE IF NOT EXISTS personal_activities (
    id              SERIAL PRIMARY KEY,
    company_id      BIGINT NOT NULL DEFAULT 1,
    description     TEXT,
    root_name       TEXT,
    activity_date   TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'OPEN',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_personal_activities_status ON personal_activities (status);

INSERT INTO personal_activities (description, root_name, activity_date, status)
SELECT 'Preparare report vendite', 'Vendite', now() + interval '2 days', 'OPEN'
WHERE NOT EXISTS (SELECT 1 FROM personal_activities WHERE description = 'Preparare report vendite');
