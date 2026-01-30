package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ptalbayeva/go-musthave-diploma/internal/client"
)

// Структуры для разбора ответа
type Order struct {
	Number  string `json:"number"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}

type OrdersResponse struct {
	Orders []Order `json:"orders"`
}

// AwaitOrderProcessed ожидает, пока заказ не будет обработан с нужным статусом
func AwaitOrderProcessed(orderNum string, expectedAccrual int, client *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("Не удалось дождаться окончания расчета начисления")
		case <-ticker.C:
			var orders OrdersResponse

			// Формируем URL для запроса
			url := fmt.Sprintf("%s/api/user/orders", client.BaseURL)

			// Создаем новый GET-запрос
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return fmt.Errorf("Ошибка при создании запроса: %v", err)
			}

			// Устанавливаем контекст запроса
			req = req.WithContext(ctx)

			// Выполняем запрос
			resp, err := client.HttpClient.Do(req)
			if err != nil {
				return fmt.Errorf("Ошибка при попытке сделать запрос на получение статуса расчета начисления: %v", err)
			}
			defer resp.Body.Close()

			// Проверяем статус ответа
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
				return fmt.Errorf("Несоответствие статус кода ответа ожидаемому: %s %s", resp.Status, req.URL)
			}

			// Проверка на Content-Type
			if !containsJSONContentType(resp) {
				return fmt.Errorf("Заголовок ответа Content-Type не содержит 'application/json'")
			}

			// Разбираем тело ответа
			if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
				return fmt.Errorf("Ошибка при разборе тела ответа: %v", err)
			}

			// Проверяем статус заказа
			if resp.StatusCode == http.StatusOK && len(orders.Orders) > 0 {
				o := orders.Orders[0]
				if o.Number == orderNum && o.Status == "PROCESSED" && o.Accrual == expectedAccrual {
					return nil // Успешно получен нужный статус и начисление
				}
			}
		}
	}
}

// Проверка, что Content-Type в ответе - application/json
func containsJSONContentType(resp *http.Response) bool {
	return resp.Header.Get("Content-Type") == "application/json"
}
