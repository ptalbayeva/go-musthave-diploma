package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client interface {
	GetOrder(orderNum string) (*OrderResponse, int, int, error)
}

type OrderResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type client struct {
	BaseURL string
	Client  *http.Client
}

func NewClient(baseURL string) *client {
	return &client{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetOrder Получение статуса заказа
func (c *client) GetOrder(orderNum string) (*OrderResponse, int, int, error) {
	url, _ := url.JoinPath(c.BaseURL, "api", "orders", orderNum)
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
