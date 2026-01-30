package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/client"
	"github.com/ptalbayeva/go-musthave-diploma/internal/service"
)

func AwaitOrderProcessed(
	ctx context.Context,
	orderID string,
	userID uuid.UUID,
	accrual *client.Client,
	loyalty *service.LoyaltyService,
) {

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res, err := accrual.GetOrder(orderID)
			if err != nil {
				continue
			}

			switch res.Status {
			case "PROCESSING":
				continue

			case "INVALID":
				_ = loyalty.UpdateOrderStatus(ctx, orderID, "INVALID")
				return

			case "PROCESSED":
				if res.Accrual == nil {
					return
				}

				_ = loyalty.UpdateOrderStatus(ctx, orderID, "PROCESSED")
				_ = loyalty.AddPoints(ctx, userID, orderID, int(*res.Accrual))
				return
			}
		}
	}
}
