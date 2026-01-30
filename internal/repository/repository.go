package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
)

var (
	ErrOrderExistsUser  = errors.New("order already exists for this user")
	ErrOrderExistsOther = errors.New("order already exists for another user")
	ErrNotFound         = errors.New("order not found")
	ErrConflict         = errors.New("login already exists")
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
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	UpdateOrderPoints(ctx context.Context, orderID string, points float32) error
	UpdateWithdrawalStatus(ctx context.Context, orderID string, status string) error
	GetByOrderID(ctx context.Context, orderID string) (models.Order, error)
	OrderExists(ctx context.Context, orderID string) (bool, uuid.UUID, error)
}

type LoyaltyRepository interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (float32, error)
	AddPoints(ctx context.Context, userID uuid.UUID, points float32) error
	SubtractPoints(ctx context.Context, userID uuid.UUID, points float32) error
	AddTransaction(ctx context.Context, userID uuid.UUID, orderID string, points float32, transactionType string) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error)
	GetWithdrawn(ctx context.Context, userID uuid.UUID) (float32, error)
}
