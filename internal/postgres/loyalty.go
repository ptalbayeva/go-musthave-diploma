package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
)

type loyaltyRepo struct {
	db *sql.DB
}

func NewLoyaltyRepository(db *sql.DB) repository.LoyaltyRepository {
	return &loyaltyRepo{db: db}
}

func (r *loyaltyRepo) GetBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	var balance int
	err := r.db.QueryRowContext(ctx, "SELECT balance FROM loyalty_points WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (r *loyaltyRepo) AddPoints(ctx context.Context, userID uuid.UUID, points int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE loyalty_points SET balance = balance + $1 WHERE user_id = $2", points, userID)
	return err
}

func (r *loyaltyRepo) SubtractPoints(ctx context.Context, userID uuid.UUID, points int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE loyalty_points SET balance = balance - $1 WHERE user_id = $2", points, userID)
	return err
}

func (r *loyaltyRepo) AddTransaction(ctx context.Context, userID uuid.UUID, orderID string, points int, transactionType string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO transactions (user_id, order_id, points, transaction_type, created_at) VALUES ($1, $2, $3, $4, NOW())", userID, orderID, points, transactionType)
	return err
}

func (r *loyaltyRepo) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, user_id, order_id, points, transaction_type, created_at FROM transactions WHERE user_id = $1 AND transaction_type = 'WITHDRAWAL'", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.OrderID, &t.Points, &t.TransactionType, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}
