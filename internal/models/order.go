package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	OrderID           string    `json:"order_id"`
	Status            string    `json:"status"`
	PointsAccumulated int       `json:"points_accumulated"`
	WithdrawalStatus  string    `json:"withdrawal_status"`
	CreatedAt         time.Time `json:"created_at"`
}
