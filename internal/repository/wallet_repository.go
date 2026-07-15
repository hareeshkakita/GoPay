package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	dbsqlc "github.com/hareeshkakita/gopay/internal/db/sqlc/gen"
)

type WalletRepository struct {
	db      *sql.DB
	queries *dbsqlc.Queries
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{
		db:      db,
		queries: dbsqlc.New(db),
	}
}

func (r *WalletRepository) CreateWallet(ctx context.Context, ownerID uuid.UUID, currency string) (dbsqlc.Wallet, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return dbsqlc.Wallet{}, err
	}
	defer tx.Rollback()

	q := r.queries.WithTx(tx)

	wallet, err := q.CreateWallet(ctx, dbsqlc.CreateWalletParams{
		ID:       uuid.New(),
		OwnerID:  ownerID,
		Currency: currency,
	})
	if err != nil {
		return dbsqlc.Wallet{}, err
	}

	_, err = q.CreateWalletBalance(ctx, dbsqlc.CreateWalletBalanceParams{
		ID:               uuid.New(),
		WalletID:         wallet.ID,
		AvailableBalance: 0,
		PendingBalance:   0,
		Currency:         currency,
	})
	if err != nil {
		return dbsqlc.Wallet{}, err
	}

	if err := tx.Commit(); err != nil {
		return dbsqlc.Wallet{}, err
	}

	return wallet, nil
}

func (r *WalletRepository) GetWalletByID(ctx context.Context, walletID uuid.UUID) (dbsqlc.Wallet, error) {
	return r.queries.GetWalletByID(ctx, walletID)
}

func (r *WalletRepository) GetBalanceByWalletID(ctx context.Context, walletID uuid.UUID) (dbsqlc.WalletBalance, error) {
	return r.queries.GetBalanceByWalletID(ctx, walletID)
}

func (r *WalletRepository) UpdateBalanceByWalletID(ctx context.Context, updatedBalance dbsqlc.ApplyNewBalanceParams) (dbsqlc.WalletBalance, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return dbsqlc.WalletBalance{}, err
	}
	defer tx.Rollback()

	q := r.queries.WithTx(tx)

	currentBalance, err := q.GetBalanceByWalletIDForUpdate(ctx, updatedBalance.WalletID)

	if err != nil {
		return dbsqlc.WalletBalance{}, err
	}

	if updatedBalance.AvailableBalance < 0 && -1*updatedBalance.AvailableBalance > currentBalance.AvailableBalance {
		return dbsqlc.WalletBalance{}, errors.New("low balance")
	}

	balance, err := q.ApplyNewBalance(ctx, updatedBalance)

	if err != nil {
		return dbsqlc.WalletBalance{}, err
	}

	if err := tx.Commit(); err != nil {
		return dbsqlc.WalletBalance{}, err
	}

	return balance, nil
}

func (r *WalletRepository) TransferAmount(ctx context.Context, source uuid.UUID, target uuid.UUID, amount int64) (dbsqlc.WalletBalance, dbsqlc.WalletBalance, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}
	defer tx.Rollback()

	q := r.queries.WithTx(tx)

	first, second := lockOrder(source, target)

	currentBalanceFirst, err := q.GetBalanceByWalletIDForUpdate(ctx, first)

	if first == source && amount > currentBalanceFirst.AvailableBalance {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, errors.New("low balance")
	}

	currentBalanceSecond, err := q.GetBalanceByWalletIDForUpdate(ctx, second)

	if second == source && amount > currentBalanceSecond.AvailableBalance {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, errors.New("low balance")
	}

	updatedBalanceSource, err := q.ApplyNewBalance(ctx, dbsqlc.ApplyNewBalanceParams{
		WalletID:         source,
		AvailableBalance: -1 * amount,
	})

	updatedBalanceTarget, err := q.ApplyNewBalance(ctx, dbsqlc.ApplyNewBalanceParams{
		WalletID:         target,
		AvailableBalance: amount,
	})

	if err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}

	if err := tx.Commit(); err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}

	return updatedBalanceSource, updatedBalanceTarget, nil
}

func lockOrder(source uuid.UUID, target uuid.UUID) (uuid.UUID, uuid.UUID) {
	if source.ID() > target.ID() {
		return source, target
	}
	return target, source
}
