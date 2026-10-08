-- Soft delete for countries, so DELETE can be undone by PATCH /countries/{id}/recover.

ALTER TABLE countries
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
