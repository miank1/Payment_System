package model

import "time"

type Payment struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Amount         int64     `json:"amount"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Payment status constants
const (
	StatusCreated    = "CREATED"
	StatusProcessing = "PROCESSING"
	StatusSuccess    = "SUCCESS"
	StatusFailed     = "FAILED"
)
