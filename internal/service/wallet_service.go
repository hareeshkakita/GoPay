package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	dbsqlc "github.com/hareeshkakita/gopay/internal/db/sqlc/gen"
	"github.com/hareeshkakita/gopay/internal/repository"
)

type WalletService struct {
	db      *sql.DB
	queries *dbsqlc.Queries

	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	ledgerRepo      repository.LedgerRepository
}

func NewWalletService(db *sql.DB, walletRepo repository.WalletRepository, transactionRepo repository.TransactionRepository, ledgerRepo repository.LedgerRepository) *WalletService {
	return &WalletService{
		db:              db,
		queries:         dbsqlc.New(db),
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		ledgerRepo:      ledgerRepo}
}

func (s *WalletService) CreateWallet(ctx context.Context, ownerID uuid.UUID, currency string) (dbsqlc.Wallet, error) {
	if currency == "" {
		currency = "INR"
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dbsqlc.Wallet{}, err
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)

	newWallet, err := s.walletRepo.CreateWallet(ctx, q, ownerID, currency)

	if err := tx.Commit(); err != nil {
		return dbsqlc.Wallet{}, err
	}

	return newWallet, nil
}

func (s *WalletService) GetWallet(ctx context.Context, walletID uuid.UUID) (dbsqlc.Wallet, error) {
	return s.walletRepo.GetWalletByID(ctx, walletID)
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (dbsqlc.WalletBalance, error) {
	return s.walletRepo.GetBalanceByWalletID(ctx, walletID)
}

func (s *WalletService) UpdateBalanceByWalletID(ctx context.Context, updatedBalance dbsqlc.ApplyNewBalanceParams) (dbsqlc.WalletBalance, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dbsqlc.WalletBalance{}, err
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)

	transactionID, err := uuid.NewUUID()
	transactionType := "CREDIT"
	if updatedBalance.AvailableBalance < 0 {
		transactionType = "DEBIT"
	}

	if err != nil {
		return dbsqlc.WalletBalance{}, err
	}

	err = s.transactionRepo.Create(ctx, q, dbsqlc.CreateTransactionParams{
		ID:     transactionID,
		Type:   transactionType,
		Status: "Done",
	})

	if err != nil {
		return dbsqlc.WalletBalance{}, err
	}

	ledgerID, err := uuid.NewUUID()

	err = s.ledgerRepo.Create(ctx, q, dbsqlc.CreateLedgerEntryParams{
		ID:            ledgerID,
		TransactionID: transactionID,
		WalletID:      updatedBalance.WalletID,
		EntryType:     transactionType,
		Amount:        updatedBalance.AvailableBalance,
	})

	if err != nil {
		return dbsqlc.WalletBalance{}, err
	}

	newBalance, err := s.walletRepo.UpdateBalanceByWalletID(ctx, q, updatedBalance)

	if err != nil {
		return dbsqlc.WalletBalance{}, err
	}

	if err := tx.Commit(); err != nil {
		return dbsqlc.WalletBalance{}, err
	}

	return newBalance, nil
}

func (s *WalletService) TransferAmount(ctx context.Context, source uuid.UUID, target uuid.UUID, amount int64) (dbsqlc.WalletBalance, dbsqlc.WalletBalance, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}
	defer tx.Rollback()

	q := s.queries.WithTx(tx)

	transactionID, err := uuid.NewUUID()
	transactionType := "TRANSFER"

	if err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}

	err = s.transactionRepo.Create(ctx, q, dbsqlc.CreateTransactionParams{
		ID:     transactionID,
		Type:   transactionType,
		Status: "Done",
	})

	if err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}

	ledgerSourceID, err := uuid.NewUUID()

	err = s.ledgerRepo.Create(ctx, q, dbsqlc.CreateLedgerEntryParams{
		ID:            ledgerSourceID,
		TransactionID: transactionID,
		WalletID:      source,
		EntryType:     "DEBIT",
		Amount:        -1 * amount,
	})

	ledgerTargetID, err := uuid.NewUUID()

	err = s.ledgerRepo.Create(ctx, q, dbsqlc.CreateLedgerEntryParams{
		ID:            ledgerTargetID,
		TransactionID: transactionID,
		WalletID:      target,
		EntryType:     "CREDIT",
		Amount:        amount,
	})

	if err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}

	updatedSourceBalance, updatedTargetBalance, err := s.walletRepo.TransferAmount(ctx, q, source, target, amount)

	if err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}

	if err := tx.Commit(); err != nil {
		return dbsqlc.WalletBalance{}, dbsqlc.WalletBalance{}, err
	}

	return updatedSourceBalance, updatedTargetBalance, nil

}
