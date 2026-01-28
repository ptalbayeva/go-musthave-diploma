package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"

	"database/sql"
)

// Реализация UsersRepository
type usersRepo struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) repository.UsersRepository {
	return &usersRepo{db: db}
}

// Create создает нового пользователя
func (r *usersRepo) Create(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
	userID := uuid.New()
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)", userID, email, passwordHash)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

// GetByEmail находит пользователя по email
func (r *usersRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, "SELECT id, email, password_hash FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return &models.User{}, errors.New("user not found")
	}
	return &user, nil
}
