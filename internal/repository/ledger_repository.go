package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	dbsqlc "github.com/hareeshkakita/gopay/internal/db/sqlc/gen"
)

type LedgerRepository interface {
	Create(ctx context.Context, q *dbsqlc.Queries, p dbsqlc.CreateLedgerEntryParams) error

	GetByID(ctx context.Context, id uuid.UUID) ([]dbsqlc.LedgerEntry, error)
}

type ledgerRepository struct {
	queries *dbsqlc.Queries
}

func NewLedgerRepository(db *sql.DB) LedgerRepository {
	return &ledgerRepository{
		queries: dbsqlc.New(db),
	}
}

func (r *ledgerRepository) Create(ctx context.Context, q *dbsqlc.Queries, p dbsqlc.CreateLedgerEntryParams) error {
	return q.CreateLedgerEntry(ctx, p)
}

func (r *ledgerRepository) GetByID(ctx context.Context, id uuid.UUID) ([]dbsqlc.LedgerEntry, error) {
	ledgerEntries, err := r.queries.GetLedgerEntriesByTransactionID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ledgerEntries, nil
}
