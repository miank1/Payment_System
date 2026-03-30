CREATE TABLE IF NOT EXISTS payments (
    id                TEXT PRIMARY KEY,
    user_id           TEXT NOT NULL,
    amount            BIGINT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'CREATED',
    idempotency_key   TEXT UNIQUE NOT NULL,
    created_at        TIMESTAMP DEFAULT NOW(),
    updated_at        TIMESTAMP DEFAULT NOW()
);