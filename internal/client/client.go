package client

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// RegisterOrder Регистрация заказа в accrual
func (c *Client) RegisterOrder(orderNum string) error {
	url := fmt.Sprintf("%s/api/orders/%s", c.BaseURL, orderNum)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted, http.StatusConflict:
		return nil
	default:
		return fmt.Errorf("accrual register error: %s", resp.Status)
	}
}

// GetOrder Получение статуса заказа
func (c *Client) GetOrder(orderNum string) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.BaseURL, orderNum)

	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accrual get error: %s", resp.Status)
	}

	var res OrderResponse
	return &res, json.NewDecoder(resp.Body).Decode(&res)
}
