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

// Получить баланс пользователя
func (r *loyaltyRepo) GetBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	var balance int
	err := r.db.QueryRowContext(ctx, "SELECT balance FROM loyalty_points WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Нет записей для пользователя, баланс = 0
		}
		return 0, err
	}
	return balance, nil
}

// Добавить баллы пользователю
func (r *loyaltyRepo) AddPoints(ctx context.Context, userID uuid.UUID, points int) error {
	// Проверка, если записи нет — создаем её
	_, err := r.db.ExecContext(ctx, "INSERT INTO loyalty_points (user_id, balance) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET balance = loyalty_points.balance + $2", userID, points)
	return err
}

// Вычесть баллы у пользователя
func (r *loyaltyRepo) SubtractPoints(ctx context.Context, userID uuid.UUID, points int) error {
	// Проверка, если записи нет — ошибка
	_, err := r.db.ExecContext(ctx, "UPDATE loyalty_points SET balance = balance - $1 WHERE user_id = $2 AND balance >= $1", points, userID)
	if err != nil {
		return err
	}
	return nil
}

// Добавить транзакцию
func (r *loyaltyRepo) AddTransaction(ctx context.Context, userID uuid.UUID, orderID string, points int, transactionType string) error {
	// Генерация уникального ID транзакции
	transactionID := uuid.New()

	// Вставка транзакции в базу данных
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO transactions (id, user_id, order_id, points, type, processed_at) VALUES ($1, $2, $3, $4, $5, NOW())",
		transactionID, userID, orderID, points, transactionType,
	)
	return err
}

func (r *loyaltyRepo) GetWithdrawn(ctx context.Context, userID uuid.UUID) (int, error) {
	var withdrawn int
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(points),0) FROM transactions WHERE user_id=$1`, userID).Scan(&withdrawn)
	return withdrawn, err
}

// Получить все транзакции вывода средств
func (r *loyaltyRepo) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, order_id, user_id, points, type, processed_at FROM transactions WHERE user_id = $1 AND type = 'WITHDRAWAL'", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.OrderID, &transaction.UserID, &transaction.Points, &transaction.Type, &transaction.ProcessedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
