package models

import (
	"time"

	"github.com/google/uuid"
)

// Модель транзакции
type Transaction struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	OrderID         string    `json:"order_id"`
	Points          int       `json:"points"`
	TransactionType string    `json:"transaction_type"` // ACCRUAL, WITHDRAWAL
	CreatedAt       time.Time `json:"created_at"`
}
