-- PaymentTerm

CREATE TABLE IF NOT EXISTS payment_terms (
    id              SERIAL PRIMARY KEY,
    id_payment_type INTEGER,
    description     TEXT NOT NULL,
    end_of_month    BOOLEAN NOT NULL DEFAULT false,
    extra_days      INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS payment_term_positions (
    id              SERIAL PRIMARY KEY,
    payment_term_id INTEGER NOT NULL REFERENCES payment_terms(id) ON DELETE CASCADE,
    days            INTEGER NOT NULL DEFAULT 0,
    percentage      DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_order  INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_payment_term_positions_term ON payment_term_positions (payment_term_id);

INSERT INTO payment_terms (id_payment_type, description, end_of_month, extra_days) VALUES
    (1, 'Pagamento a 30 giorni', false, 30),
    (1, 'Pagamento a 60 giorni fine mese', true, 0);

INSERT INTO payment_term_positions (payment_term_id, days, percentage, position_order)
SELECT pt.id, 30, 100, 1
FROM payment_terms pt
WHERE pt.description = 'Pagamento a 30 giorni'
  AND NOT EXISTS (
      SELECT 1 FROM payment_term_positions p WHERE p.payment_term_id = pt.id
  );

INSERT INTO payment_term_positions (payment_term_id, days, percentage, position_order)
SELECT pt.id, 60, 100, 1
FROM payment_terms pt
WHERE pt.description = 'Pagamento a 60 giorni fine mese'
  AND NOT EXISTS (
      SELECT 1 FROM payment_term_positions p WHERE p.payment_term_id = pt.id
  );
