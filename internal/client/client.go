package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OrderResponse представляет структуру ответа от API после регистрации заказа
type OrderResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// Client представляет собой экземпляр клиента для общения с API
type Client struct {
	HttpClient *http.Client
	BaseURL    string
}

// NewClient создает новый экземпляр клиента
func NewClient(baseURL string) *Client {
	return &Client{
		HttpClient: &http.Client{},
		BaseURL:    baseURL,
	}
}

// RegisterOrder регистрирует новый номер заказа через API
func (c *Client) RegisterOrder(orderNum string) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.BaseURL, orderNum)

	reqBody := []byte(orderNum)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while processing order: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Order registered successfully")
	} else if resp.StatusCode == http.StatusAccepted {
		fmt.Println("Order accepted")
	} else {
		return nil, fmt.Errorf("error while processing order, status: %s, response: %s", resp.Status, string(respBody))
	}

	return &orderResp, nil
}
