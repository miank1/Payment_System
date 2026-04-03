package ledger

import "time"

type Entry struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"payment_id"`
	Type      string    `json:"type"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	TypeDebit  = "DEBIT"
	TypeCredit = "CREDIT"
)
