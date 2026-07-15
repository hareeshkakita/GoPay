-- internal/db/sqlc/queries.sql

-- name: CreateWallet :one
INSERT INTO wallets (id, owner_id, currency)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetWalletByID :one
SELECT *
FROM wallets
WHERE id = $1
LIMIT 1;

-- name: CreateWalletBalance :one
INSERT INTO wallet_balances (id, wallet_id, available_balance, pending_balance, currency)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetBalanceByWalletID :one
SELECT *
FROM wallet_balances
WHERE wallet_id = $1
LIMIT 1;

-- name: UpdateBalance :one
UPDATE wallet_balances
SET available_balance = $2,
    pending_balance = $3,
    updated_at = NOW()
WHERE wallet_id = $1
RETURNING *;

-- name: ApplyNewBalance :one
UPDATE wallet_balances
SET available_balance = available_balance + $2,
    updated_at = NOW()
WHERE wallet_id = $1
RETURNING *;

-- name: GetBalanceByWalletIDForUpdate :one
SELECT *
FROM wallet_balances
WHERE wallet_id = $1
FOR UPDATE
LIMIT 1;
