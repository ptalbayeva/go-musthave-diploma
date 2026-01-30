package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
)

// Логика для работы с лояльностью
type LoyaltyService struct {
	loyaltyRepo repository.LoyaltyRepository
	orderRepo   repository.OrdersRepository
}

// Конструктор для сервиса лояльности
func NewLoyaltyService(loyaltyRepo repository.LoyaltyRepository, orderRepo repository.OrdersRepository) *LoyaltyService {
	return &LoyaltyService{loyaltyRepo: loyaltyRepo, orderRepo: orderRepo}
}

// GetUserBalance Получить баланс пользователя
func (s *LoyaltyService) GetUserBalance(ctx context.Context, userID uuid.UUID) (current int, withdrawn int, err error) {
	current, err = s.loyaltyRepo.GetBalance(ctx, userID)
	if err != nil {
		return
	}
	withdrawn, err = s.loyaltyRepo.GetWithdrawn(ctx, userID)
	return
}

// Получить историю выводов средств
func (s *LoyaltyService) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	withdrawals, err := s.loyaltyRepo.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

// Получить все заказы пользователя
func (s *LoyaltyService) GetOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	orders, err := s.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// Метод для добавления баллов за заказ
func (s *LoyaltyService) AddPoints(ctx context.Context, userID uuid.UUID, orderID string, points int) error {
	// Проверяем, существует ли заказ
	order, err := s.orderRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	// Если заказ уже был обработан (статус != NEW), то не начисляем баллы
	if order.Status != "NEW" {
		return errors.New("order already processed or in an invalid state")
	}

	// Обновляем баллы в заказе
	if err := s.orderRepo.UpdateOrderPoints(ctx, orderID, points); err != nil {
		return err
	}

	// Добавляем баллы пользователю
	if err := s.loyaltyRepo.AddPoints(ctx, userID, points); err != nil {
		return err
	}

	// Добавляем транзакцию начисления
	return s.loyaltyRepo.AddTransaction(ctx, userID, orderID, points, "ACCRUAL")
}

// Метод для списания баллов
func (s *LoyaltyService) WithdrawPoints(ctx context.Context, userID uuid.UUID, points int) error {
	// Получаем текущий баланс
	balance, err := s.loyaltyRepo.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	// Проверяем, достаточно ли баллов
	if balance < points {
		return errors.New("insufficient balance")
	}

	// Вычитаем баллы из баланса пользователя
	if err := s.loyaltyRepo.SubtractPoints(ctx, userID, points); err != nil {
		return err
	}

	// Добавляем транзакцию вывода средств
	transaction := models.Transaction{
		OrderID:     "withdrawal", // Используем специальный идентификатор для вывода
		UserID:      userID,
		Points:      -points, // Баллы будут отрицательными для вывода
		Type:        "WITHDRAWAL",
		ProcessedAt: time.Now().Format(time.RFC3339),
	}

	// Записываем транзакцию
	return s.loyaltyRepo.AddTransaction(ctx, userID, transaction.OrderID, transaction.Points, transaction.Type)
}

// OrderExists Метод для проверки существования заказа в сервисе лояльности
func (s *LoyaltyService) OrderExists(ctx context.Context, orderID string) (bool, uuid.UUID, error) {
	return s.orderRepo.OrderExists(ctx, orderID)
}

// CreateOrder Метод для создания нового заказа и начисления баллов
func (s *LoyaltyService) CreateOrder(ctx context.Context, userID uuid.UUID, orderID string) error {
	const status = "NEW"

	_, err := s.orderRepo.CreateOrder(ctx, userID, orderID, status)
	if err != nil {
		return err
	}

	return nil
}

// Метод для обновления статуса заказа
func (s *LoyaltyService) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	// Обновляем статус заказа
	if err := s.orderRepo.UpdateWithdrawalStatus(ctx, orderID, status); err != nil {
		return err
	}

	order, err := s.orderRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	// Добавляем транзакцию
	transaction := models.Transaction{
		OrderID:     order.OrderID,
		UserID:      order.UserID,
		Points:      order.PointsAccumulated,
		Type:        "STATUS_UPDATE",
		ProcessedAt: time.Now().Format(time.RFC3339),
	}

	// Записываем транзакцию
	return s.loyaltyRepo.AddTransaction(ctx, transaction.UserID, transaction.OrderID, transaction.Points, transaction.Type)
}
