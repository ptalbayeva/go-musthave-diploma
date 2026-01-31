package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
	"github.com/ptalbayeva/go-musthave-diploma/internal/service"
	"github.com/stretchr/testify/require"
)

type mockLoyaltyRepo struct {
	GetCurrentFn     func(ctx context.Context, userID uuid.UUID) (float32, error)
	GetWithdrawnFn   func(ctx context.Context, userID uuid.UUID) (float32, error)
	GetWithdrawalsFn func(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error)
	AddTransactionFn func(ctx context.Context, userID uuid.UUID, orderID string, points float32, t string) error
	AddPointsFn      func(ctx context.Context, userID uuid.UUID, points float32)
	SubtractPointsFn func(ctx context.Context, userID uuid.UUID, points float32) error
}

func (m *mockLoyaltyRepo) AddPoints(ctx context.Context, userID uuid.UUID, points float32) error {
	return m.AddPoints(ctx, userID, points)
}

func (m *mockLoyaltyRepo) SubtractPoints(ctx context.Context, userID uuid.UUID, points float32) error {
	return m.SubtractPoints(ctx, userID, points)
}

func (m *mockLoyaltyRepo) GetCurrent(ctx context.Context, userID uuid.UUID) (float32, error) {
	return m.GetCurrentFn(ctx, userID)
}

func (m *mockLoyaltyRepo) GetWithdrawn(ctx context.Context, userID uuid.UUID) (float32, error) {
	return m.GetWithdrawnFn(ctx, userID)
}

func (m *mockLoyaltyRepo) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	return m.GetWithdrawalsFn(ctx, userID)
}

func (m *mockLoyaltyRepo) AddTransaction(
	ctx context.Context,
	userID uuid.UUID,
	orderID string,
	points float32,
	t string,
) error {
	return m.AddTransactionFn(ctx, userID, orderID, points, t)
}

type mockOrdersRepo struct {
	GetOrdersByUserIDFn      func(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	GetByOrderIDFn           func(ctx context.Context, orderID string) (models.Order, error)
	UpdateOrderPointsFn      func(ctx context.Context, orderID string, points float32) error
	OrderExistsFn            func(ctx context.Context, orderID string) (bool, uuid.UUID, error)
	CreateOrderFn            func(ctx context.Context, userID uuid.UUID, orderID, status string) (uuid.UUID, error)
	UpdateWithdrawalStatusFn func(ctx context.Context, orderID, status string) error
}

func (m *mockOrdersRepo) GetByOrderID(ctx context.Context, orderID string) (models.Order, error) {
	return m.GetByOrderIDFn(ctx, orderID)
}

func (m *mockOrdersRepo) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return m.GetOrdersByUserIDFn(ctx, userID)
}

func (m *mockOrdersRepo) UpdateOrderPoints(ctx context.Context, orderID string, points float32) error {
	return m.UpdateOrderPointsFn(ctx, orderID, points)
}

func (m *mockOrdersRepo) OrderExists(ctx context.Context, orderID string) (bool, uuid.UUID, error) {
	return m.OrderExistsFn(ctx, orderID)
}

func (m *mockOrdersRepo) CreateOrder(ctx context.Context, userID uuid.UUID, orderID, status string) (uuid.UUID, error) {
	return m.CreateOrderFn(ctx, userID, orderID, status)
}

func (m *mockOrdersRepo) UpdateWithdrawalStatus(ctx context.Context, orderID, status string) error {
	return m.UpdateWithdrawalStatusFn(ctx, orderID, status)
}

// TestLoyaltyService_WithdrawPoints проверка списания с баланса пользователя
func TestLoyaltyService_WithdrawPoints(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		current       float32
		points        float32
		expectedError error
	}{
		{
			name:          "success",
			current:       100,
			points:        50,
			expectedError: nil,
		},
		{
			name:          "insufficient balance",
			current:       10,
			points:        50,
			expectedError: service.ErrInsufficientBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loyaltyRepo := &mockLoyaltyRepo{
				GetCurrentFn: func(ctx context.Context, id uuid.UUID) (float32, error) {
					return tt.current, nil
				},
				AddTransactionFn: func(ctx context.Context, id uuid.UUID, orderID string, points float32, tsType string) error {
					require.Equal(t, "WITHDRAWAL", tsType)
					return nil
				},
			}

			svc := service.NewLoyaltyService(loyaltyRepo, nil)

			err := svc.WithdrawPoints(context.Background(), userID, "order-1", tt.points)

			if tt.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.expectedError)
			}
		})
	}
}

// TestLoyaltyService_GetUserBalance проверка баланса пользователя
func TestLoyaltyService_GetUserBalance(t *testing.T) {
	loyaltyRepo := &mockLoyaltyRepo{
		GetCurrentFn: func(ctx context.Context, userID uuid.UUID) (float32, error) {
			return 100, nil
		},
		GetWithdrawnFn: func(ctx context.Context, userID uuid.UUID) (float32, error) {
			return 40, nil
		},
	}

	svc := service.NewLoyaltyService(loyaltyRepo, nil)

	current, withdrawn, err := svc.GetUserBalance(context.Background(), uuid.New())

	require.NoError(t, err)
	require.Equal(t, float32(100), current)
	require.Equal(t, float32(40), withdrawn)
}

// TestLoyaltyService_AddPoints проверка добавленияы бонусов
func TestLoyaltyService_AddPoints(t *testing.T) {
	userID := uuid.New()

	orderRepo := &mockOrdersRepo{
		GetByOrderIDFn: func(ctx context.Context, orderID string) (models.Order, error) {
			return models.Order{OrderID: orderID}, nil
		},
		UpdateOrderPointsFn: func(ctx context.Context, orderID string, points float32) error {
			return nil
		},
	}

	loyaltyRepo := &mockLoyaltyRepo{
		AddTransactionFn: func(ctx context.Context, id uuid.UUID, orderID string, points float32, tsType string) error {
			require.Equal(t, "ACCRUAL", tsType)
			return nil
		},
	}

	svc := service.NewLoyaltyService(loyaltyRepo, orderRepo)

	err := svc.AddPoints(context.Background(), userID, "order-1", 100)

	require.NoError(t, err)
}
