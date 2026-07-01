package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrderID uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Amount  float64   `gorm:"not null" json:"amount"`
	Status  string    `gorm:"not null;default:PENDING" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
