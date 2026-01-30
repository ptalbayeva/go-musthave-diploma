package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type OrderResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type Client struct {
	BaseURL string
	Client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetOrder Получение статуса заказа
func (c *Client) GetOrder(orderNum string) (*OrderResponse, int, int, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.BaseURL, orderNum)

	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var res OrderResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return nil, resp.StatusCode, 0, err
		}

		fmt.Println(&res)
		return &res, resp.StatusCode, 0, nil

	case http.StatusNoContent:
		return nil, resp.StatusCode, 0, nil

	case http.StatusTooManyRequests:
		retryAfter := 60
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if v, err := strconv.Atoi(ra); err == nil {
				retryAfter = v
			}
		}
		return nil, resp.StatusCode, retryAfter, nil

	case http.StatusInternalServerError:
		return nil, resp.StatusCode, 0, fmt.Errorf("server error: %s", resp.Status)

	default:
		return nil, resp.StatusCode, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
