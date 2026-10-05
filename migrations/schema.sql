CREATE TABLE IF NOT EXISTS orders (
    id              SERIAL PRIMARY KEY,
    idempotency_key TEXT UNIQUE NOT NULL,
    origin          TEXT NOT NULL,
    destination     TEXT NOT NULL,
    weight_kg       DOUBLE PRECISION NOT NULL,
    carrier         TEXT NOT NULL,
    price_bdt       DOUBLE PRECISION NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
