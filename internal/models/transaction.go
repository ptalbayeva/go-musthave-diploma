package models

import "github.com/google/uuid"

type Transaction struct {
	ID          uuid.UUID `json:"id"`
	OrderID     string    `json:"order_id"`
	UserID      uuid.UUID `json:"user_id"`
	Points      int       `json:"points"`
	Type        string    `json:"type"` // "ACCRUAL", "WITHDRAWAL", "STATUS_UPDATE"
	ProcessedAt string    `json:"processed_at"`
}
