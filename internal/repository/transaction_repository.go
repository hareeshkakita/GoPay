package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	dbsqlc "github.com/hareeshkakita/gopay/internal/db/sqlc/gen"
)

type TransactionRepository interface {
	Create(ctx context.Context, q *dbsqlc.Queries, p dbsqlc.CreateTransactionParams) error

	UpdateStatus(ctx context.Context, q *dbsqlc.Queries, transactionID uuid.UUID, status string) error

	GetByID(ctx context.Context, id uuid.UUID) (*dbsqlc.Transaction, error)
}

type transactionRepository struct {
	queries *dbsqlc.Queries
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{
		queries: dbsqlc.New(db),
	}
}

func (r *transactionRepository) Create(
	ctx context.Context, q *dbsqlc.Queries,
	p dbsqlc.CreateTransactionParams,
) error {

	return q.CreateTransaction(ctx, dbsqlc.CreateTransactionParams{
		ID:     p.ID,
		Type:   p.Type,
		Status: p.Status,
	})
}

func (r *transactionRepository) UpdateStatus(
	ctx context.Context, q *dbsqlc.Queries,
	id uuid.UUID,
	status string,
) error {

	return q.UpdateTransactionStatus(ctx,
		dbsqlc.UpdateTransactionStatusParams{
			ID:     id,
			Status: status,
		})
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*dbsqlc.Transaction, error) {
	tx, err := r.queries.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}
