package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
)

type LoyaltyService struct {
	loyaltyRepo repository.LoyaltyRepository
	orderRepo   repository.OrdersRepository
}

func NewLoyaltyService(loyaltyRepo repository.LoyaltyRepository, orderRepo repository.OrdersRepository) *LoyaltyService {
	return &LoyaltyService{loyaltyRepo: loyaltyRepo, orderRepo: orderRepo}
}

func (s *LoyaltyService) GetBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	// Получаем баланс лояльности пользователя
	balance, err := s.loyaltyRepo.GetBalance(ctx, userID)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (s *LoyaltyService) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	// Получаем историю выводов средств
	withdrawals, err := s.loyaltyRepo.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

func (s *LoyaltyService) AddPoints(ctx context.Context, userID uuid.UUID, orderID string, points int) error {
	// Обновляем заказ с начисленными баллами
	if err := s.orderRepo.UpdateOrderPoints(ctx, orderID, points); err != nil {
		return err
	}

	// Добавляем баллы на счет пользователя
	if err := s.loyaltyRepo.AddPoints(ctx, userID, points); err != nil {
		return err
	}

	// Добавляем транзакцию начисления баллов
	return s.loyaltyRepo.AddTransaction(ctx, userID, orderID, points, "ACCRUAL")
}

func (s *LoyaltyService) WithdrawPoints(ctx context.Context, userID uuid.UUID, points int) error {
	// Получаем текущий баланс пользователя
	balance, err := s.loyaltyRepo.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if balance < points {
		return errors.New("insufficient balance")
	}

	if err := s.loyaltyRepo.SubtractPoints(ctx, userID, points); err != nil {
		return err
	}

	return nil
}

// GetOrders - получение всех заказов пользователя
func (s *LoyaltyService) GetOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	// Запрашиваем заказы из репозитория
	orders, err := s.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
