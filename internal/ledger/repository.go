package ledger

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddEntry(ctx context.Context, paymentID string, entryType string, amount int64) error {
	query := `
		INSERT INTO ledger_entries (id, payment_id, type, amount)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		uuid.NewString(),
		paymentID,
		entryType,
		amount,
	)

	return err
}
