package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
)

type Repositories struct {
	Users   UsersRepository
	Orders  OrdersRepository
	Loyalty LoyaltyRepository
}

type UsersRepository interface {
	Create(ctx context.Context, email, passwordHash string) (uuid.UUID, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type OrdersRepository interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, orderID string, status string) (uuid.UUID, error)
	GetByOrderID(ctx context.Context, orderID string) (models.Order, error)
	UpdateOrderPoints(ctx context.Context, orderID string, points int) error
	UpdateWithdrawalStatus(ctx context.Context, orderID string, status string) error
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
}

type LoyaltyRepository interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (int, error)
	AddPoints(ctx context.Context, userID uuid.UUID, points int) error
	SubtractPoints(ctx context.Context, userID uuid.UUID, points int) error
	AddTransaction(ctx context.Context, userID uuid.UUID, orderID string, points int, transactionType string) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) // Получение всех выводов
}
