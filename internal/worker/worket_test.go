package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/client"
	"github.com/ptalbayeva/go-musthave-diploma/internal/models"
	"github.com/ptalbayeva/go-musthave-diploma/internal/worker"
	"github.com/stretchr/testify/require"
)

// mockClient мок клиента
type mockClient struct {
	GetOrderFn func(orderID string) (*client.OrderResponse, int, int, error)
}

func (m *mockClient) GetOrder(orderID string) (*client.OrderResponse, int, int, error) {
	return m.GetOrderFn(orderID)
}

// mockLoyalty мок сервиса Loyalty
type mockLoyalty struct {
	UpdateOrderStatusFn func(ctx context.Context, orderID, status string) error
	AddPointsFn         func(ctx context.Context, userID uuid.UUID, orderID string, points float32) error
}

func (m *mockLoyalty) CreateOrder(ctx context.Context, userID uuid.UUID, orderID string) error {
	return nil
}

func (m *mockLoyalty) OrderExists(ctx context.Context, orderID string) (bool, uuid.UUID, error) {
	return false, uuid.Nil, nil
}

func (m *mockLoyalty) GetOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return []models.Order{}, nil
}

func (m *mockLoyalty) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	return []models.Transaction{}, nil
}

func (m *mockLoyalty) GetUserBalance(ctx context.Context, userID uuid.UUID) (current float32, withdrawn float32, err error) {
	return 0.0, 0.0, nil
}

func (m *mockLoyalty) WithdrawPoints(ctx context.Context, userID uuid.UUID, orderID string, points float32) error {
	return nil
}

func (m *mockLoyalty) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	if m.UpdateOrderStatusFn != nil {
		return m.UpdateOrderStatusFn(ctx, orderID, status)
	}
	return nil
}

func (m *mockLoyalty) AddPoints(ctx context.Context, userID uuid.UUID, orderID string, points float32) error {
	if m.AddPointsFn != nil {
		return m.AddPointsFn(ctx, userID, orderID, points)
	}
	return nil
}

// TestAwaitOrderProcessed
func TestAwaitOrderProcessed(t *testing.T) {
	userID := uuid.New()
	orderID := "123"

	tests := []struct {
		name         string
		getOrderFunc func(orderID string) (*client.OrderResponse, int, int, error)
		expectStatus string
		expectPoints float32
	}{
		{
			name: "processed with accrual",
			getOrderFunc: func(orderID string) (*client.OrderResponse, int, int, error) {
				v := float64(42)
				return &client.OrderResponse{
					Order:   orderID,
					Status:  "PROCESSED",
					Accrual: &v,
				}, 200, 0, nil
			},
			expectStatus: "PROCESSED",
			expectPoints: 42,
		},
		{
			name: "invalid order",
			getOrderFunc: func(orderID string) (*client.OrderResponse, int, int, error) {
				return &client.OrderResponse{
					Order:  orderID,
					Status: "INVALID",
				}, 200, 0, nil
			},
			expectStatus: "INVALID",
			expectPoints: 0,
		},
		{
			name: "order not registered (204)",
			getOrderFunc: func(orderID string) (*client.OrderResponse, int, int, error) {
				return nil, 204, 0, nil
			},
			expectStatus: "",
			expectPoints: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loyaltyCalled := ""
			pointsAdded := float32(0)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			mc := &mockClient{GetOrderFn: tt.getOrderFunc}

			ml := &mockLoyalty{
				UpdateOrderStatusFn: func(ctx context.Context, orderID, status string) error {
					loyaltyCalled = status
					return nil
				},
				AddPointsFn: func(ctx context.Context, userID uuid.UUID, orderID string, points float32) error {
					pointsAdded = points
					return nil
				},
			}

			worker.AwaitOrderProcessed(ctx, orderID, userID, mc, ml)

			require.Equal(t, tt.expectStatus, loyaltyCalled)
			require.Equal(t, tt.expectPoints, pointsAdded)
		})
	}
}
