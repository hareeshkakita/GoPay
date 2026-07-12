-- internal/db/sqlc/schema.sql

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallet_balances (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL UNIQUE REFERENCES wallets(id) ON DELETE CASCADE,
    available_balance BIGINT NOT NULL DEFAULT 0,
    pending_balance BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);