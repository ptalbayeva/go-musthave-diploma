package worker

import (
	"context"
	"fmt"
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
	ticker := time.NewTicker(1 * time.Second) // чаще опрашиваем
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context done, exiting worker")
			return
		case <-ticker.C:
			res, statusCode, retryAfter, err := accrual.GetOrder(orderID)
			if err != nil {
				fmt.Println("Error fetching order:", err)
				continue
			}

			switch statusCode {
			case 204:
				fmt.Println("Order not registered:", orderID)
				return
			case 429:
				wait := 60
				if retryAfter > 0 {
					wait = retryAfter
				}
				time.Sleep(time.Duration(wait) * time.Second)
				continue
			case 500:
				fmt.Println("Server error, retrying:", orderID)
				continue
			}

			switch res.Status {
			case "REGISTERED", "PROCESSING":
				continue
			case "INVALID":
				_ = loyalty.UpdateOrderStatus(ctx, orderID, "INVALID")
				return
			case "PROCESSED":
				points := 0
				if res.Accrual != nil {
					points = int(*res.Accrual)
				}

				_ = loyalty.UpdateOrderStatus(ctx, orderID, "PROCESSED")
				if points > 0 {
					_ = loyalty.AddPoints(ctx, userID, orderID, points)
				}
				return
			}
		}
	}
}
