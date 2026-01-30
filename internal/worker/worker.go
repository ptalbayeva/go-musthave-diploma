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
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res, statusCode, retryAfter, err := accrual.GetOrder(orderID)
			fmt.Println(res)
			if err != nil {
				fmt.Println("Error fetching order:", err)
				continue
			}

			switch statusCode {
			case 204:
				fmt.Println("Order not registered:", orderID)
				return

			case 429:
				wait := 60 // по умолчанию 60 секунд
				if retryAfter > 0 {
					wait = retryAfter
				}
				fmt.Printf("Rate limit exceeded, waiting %d seconds\n", wait)
				time.Sleep(time.Duration(wait) * time.Second)
				continue

			case 500:
				fmt.Println("Server error for order:", orderID)
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
					_ = loyalty.UpdateOrderStatus(ctx, orderID, "PROCESSED")
					return
				}

				_ = loyalty.UpdateOrderStatus(ctx, orderID, "PROCESSED")
				_ = loyalty.AddPoints(ctx, userID, orderID, int(*res.Accrual))
				return

			case "REGISTERED":
				continue
			}
		}
	}
}
