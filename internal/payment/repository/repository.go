package repository

import (
	"context"
	"database/sql"
	"payment-system/internal/payment/model"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, p *model.Payment) error {
	query := `
		INSERT INTO payments (id, user_id, amount, status, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		p.ID,
		p.UserID,
		p.Amount,
		p.Status,
		p.IdempotencyKey,
		p.CreatedAt,
		p.UpdatedAt,
	)

	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.Payment, error) {
	query := `
		SELECT id, user_id, amount, status, idempotency_key, created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	p := &model.Payment{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.UserID,
		&p.Amount,
		&p.Status,
		&p.IdempotencyKey,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *Repository) GetByIdempotencyKey(ctx context.Context, key string) (*model.Payment, error) {
	query := `
		SELECT id, user_id, amount, status, idempotency_key, created_at, updated_at
		FROM payments
		WHERE idempotency_key = $1
	`

	p := &model.Payment{}

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&p.ID,
		&p.UserID,
		&p.Amount,
		&p.Status,
		&p.IdempotencyKey,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return p, nil
}
