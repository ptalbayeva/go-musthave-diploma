package models

import "github.com/google/uuid"

// Модель баланса лояльности
type Loyalty struct {
	UserID    uuid.UUID `json:"user_id"`
	Balance   int       `json:"balance"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}
