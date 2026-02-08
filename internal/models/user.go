package models

import "github.com/google/uuid"

// Модель пользователя
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    string    `json:"created_at"`
}
