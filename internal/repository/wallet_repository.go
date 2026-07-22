package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	dbsqlc "github.com/hareeshkakita/gopay/internal/db/sqlc/gen"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context, q *dbsqlc.Queries, ownerID uuid.UUID, currency string) (dbsqlc.Wallet, error)
	GetWalletByID(ctx context.Context, walletID uuid.UUID) (dbsqlc.Wallet, error)
	GetBalanceByWalletID(ctx context.Context, walletID uuid.UUID) (dbsqlc.WalletBalance, error)
	UpdateBalanceByWalletID(ctx context.Context, q *dbsqlc.Queries, updatedBalance dbsqlc.ApplyNewBalanceParams) (dbsqlc.WalletBalance, error)
	TransferAmount(ctx context.Context, q *dbsqlc.Queries, source uuid.UUID, target uuid.UUID, amount int64) (dbsqlc.WalletBalance, dbsqlc.WalletBalance, error)
}

type walletRepository struct {
	queries *dbsqlc.Queries
}

func NewWalletRepository(db *sql.DB) WalletRepository {
	return &walletRepository{
		queries: dbsqlc.New(db),
	}
}

func (r *walletRepository) CreateWallet(ctx context.Context, q *dbsqlc.Queries, ownerID uuid.UUID, currency string) (dbsqlc.Wallet, error) {

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

	return wallet, nil
}

func (r *walletRepository) GetWalletByID(ctx context.Context, walletID uuid.UUID) (dbsqlc.Wallet, error) {
	return r.queries.GetWalletByID(ctx, walletID)
}

func (r *walletRepository) GetBalanceByWalletID(ctx context.Context, walletID uuid.UUID) (dbsqlc.WalletBalance, error) {
	return r.queries.GetBalanceByWalletID(ctx, walletID)
}

func (r *walletRepository) UpdateBalanceByWalletID(ctx context.Context, q *dbsqlc.Queries, updatedBalance dbsqlc.ApplyNewBalanceParams) (dbsqlc.WalletBalance, error) {

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

	return balance, nil
}

func (r *walletRepository) TransferAmount(ctx context.Context, q *dbsqlc.Queries, source uuid.UUID, target uuid.UUID, amount int64) (dbsqlc.WalletBalance, dbsqlc.WalletBalance, error) {

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

	return updatedBalanceSource, updatedBalanceTarget, nil
}

func lockOrder(source uuid.UUID, target uuid.UUID) (uuid.UUID, uuid.UUID) {
	if source.ID() > target.ID() {
		return source, target
	}
	return target, source
}
