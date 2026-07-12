package repository

import (
	"context"
	"database/sql"

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
