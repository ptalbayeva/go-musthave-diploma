package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
)

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) repository.OrdersRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) CreateOrder(ctx context.Context, userID uuid.UUID, orderID string, status string) (uuid.UUID, error) {
	newOrderID := uuid.New()
	_, err := r.db.ExecContext(ctx, "INSERT INTO orders (id, user_id, order_id, status, created_at) VALUES ($1, $2, $3, $4, NOW())", newOrderID, userID, orderID, status)
	if err != nil {
		return uuid.Nil, err
	}
	return newOrderID, nil
}

func (r *orderRepo) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, user_id, order_id, status, points_accumulated, withdrawal_status, created_at FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.OrderID, &order.Status, &order.PointsAccumulated, &order.WithdrawalStatus, &order.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepo) UpdateOrderPoints(ctx context.Context, orderID string, points int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE orders SET points_accumulated = $1 WHERE order_id = $2", points, orderID)
	return err
}

func (r *orderRepo) UpdateWithdrawalStatus(ctx context.Context, orderID string, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE order_id = $2", status, orderID)
	return err
}

func (r *orderRepo) GetByOrderID(ctx context.Context, orderID string) (models.Order, error) {
	var order models.Order
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, order_id, status, points_accumulated, withdrawal_status, created_at FROM orders WHERE order_id = $1", orderID).
		Scan(&order.ID, &order.UserID, &order.OrderID, &order.Status, &order.PointsAccumulated, &order.WithdrawalStatus, &order.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Order{}, nil
	}

	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

// Метод для проверки существования заказа с таким order_id
func (r *orderRepo) OrderExists(ctx context.Context, orderID string) (bool, uuid.UUID, error) {
	const query = `
		SELECT user_id
		FROM orders
		WHERE order_id = $1
	`

	var owner uuid.UUID

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(&owner)
	if err == sql.ErrNoRows {
		return false, uuid.Nil, nil
	}
	if err != nil {
		return false, uuid.Nil, err
	}

	return true, owner, nil
}
