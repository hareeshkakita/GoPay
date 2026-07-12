package service

import (
	"context"

	"github.com/google/uuid"
	dbsqlc "github.com/hareeshkakita/gopay/internal/db/sqlc/gen"
	"github.com/hareeshkakita/gopay/internal/repository"
)

type WalletService struct {
	repo *repository.WalletRepository
}

func NewWalletService(repo *repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) CreateWallet(ctx context.Context, ownerID uuid.UUID, currency string) (dbsqlc.Wallet, error) {
	if currency == "" {
		currency = "INR"
	}
	return s.repo.CreateWallet(ctx, ownerID, currency)
}

func (s *WalletService) GetWallet(ctx context.Context, walletID uuid.UUID) (dbsqlc.Wallet, error) {
	return s.repo.GetWalletByID(ctx, walletID)
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (dbsqlc.WalletBalance, error) {
	return s.repo.GetBalanceByWalletID(ctx, walletID)
}
