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

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    reference VARCHAR(100) UNIQUE,
    type VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY,

    transaction_id UUID NOT NULL
        REFERENCES transactions(id),

    wallet_id UUID NOT NULL
        REFERENCES wallets(id),

    entry_type VARCHAR(10) NOT NULL
        CHECK (entry_type IN ('DEBIT','CREDIT')),

    amount BIGINT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);