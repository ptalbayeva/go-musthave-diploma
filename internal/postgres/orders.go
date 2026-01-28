package postgres

import (
	"context"
	"database/sql"

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
	// Генерируем новый UUID для id
	newOrderID := uuid.New()

	// Вставляем новый заказ в базу данных
	_, err := r.db.ExecContext(ctx, "INSERT INTO orders (id, user_id, order_id, status, created_at) VALUES ($1, $2, $3, $4, NOW())", newOrderID, userID, orderID, status)
	if err != nil {
		return uuid.Nil, err
	}

	return newOrderID, nil
}

func (r *orderRepo) GetByOrderID(ctx context.Context, orderID string) (models.Order, error) {
	var order models.Order
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, order_id, status, points_accumulated, withdrawal_status, created_at FROM orders WHERE order_id = $1", orderID).Scan(&order.ID, &order.UserID, &order.OrderID, &order.Status, &order.PointsAccumulated, &order.WithdrawalStatus, &order.CreatedAt)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (r *orderRepo) UpdateOrderPoints(ctx context.Context, orderID string, points int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE orders SET points_accumulated = $1 WHERE order_id = $2", points, orderID)
	return err
}

func (r *orderRepo) UpdateWithdrawalStatus(ctx context.Context, orderID string, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE orders SET withdrawal_status = $1 WHERE order_id = $2", status, orderID)
	return err
}

// GetOrdersByUserID - получает все заказы пользователя по user_id
func (r *orderRepo) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	// Выполняем запрос для получения всех заказов пользователя
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

	// Проверяем, были ли ошибки при обработке строк
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
