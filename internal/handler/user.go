package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/middleware"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
	"github.com/ptalbayeva/go-musthave-diploma/internal/service"
)

type UserHandler struct {
	authService    *service.AuthService
	loyaltyService *service.LoyaltyService
	secretKey      string
	authMiddleware *middleware.AuthMiddleware
}

func NewUserHandler(authService *service.AuthService, loyaltyService *service.LoyaltyService, secretKey string) *UserHandler {
	authMiddleware := middleware.NewAuthMiddleware(secretKey)
	return &UserHandler{
		authService:    authService,
		loyaltyService: loyaltyService,
		secretKey:      secretKey,
		authMiddleware: authMiddleware,
	}
}

// Register - регистрация нового пользователя
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Регистрируем пользователя
	userID, err := h.authService.RegisterUser(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Аутентификация пользователя после регистрации
	token, err := h.authService.AuthenticateUser(r.Context(), request.Email, request.Password)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Устанавливаем токен в cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
	})

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]uuid.UUID{"user_id": userID})
}

// Login - аутентификация пользователя
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.AuthenticateUser(r.Context(), request.Email, request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Устанавливаем токен в cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
	})

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// CreateOrder - создание нового заказа и начисление баллов
func (h *UserHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Чтение тела запроса как строки
	orderIDBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	orderID := string(orderIDBytes)

	// Убираем возможные лишние пробелы в orderID
	orderID = strings.TrimSpace(orderID)

	// Проверяем, что номер заказа валиден
	if !isValidOrderID(orderID) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	// Проверяем уникальность order_id
	orderExists, err := h.loyaltyService.OrderExists(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Failed to check order existence", http.StatusInternalServerError)
		return
	}

	if orderExists {
		http.Error(w, "Order with this order_id already exists", http.StatusOK)
		return
	}

	// Создаем заказ
	_, err = h.loyaltyService.CreateOrder(r.Context(), userID, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Пример начисления баллов
	err = h.loyaltyService.AddPoints(r.Context(), userID, orderID, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Обновляем статус заказа на "PROCESSED"
	err = h.loyaltyService.UpdateOrderStatus(r.Context(), orderID, "PROCESSED")
	if err != nil {
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusAccepted)
}

// GetOrders - получение списка заказов пользователя
func (h *UserHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Используем middleware для авторизации
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.loyaltyService.GetOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type response struct {
		Number     string `json:"number"`
		Status     string `json:"status"`
		Accural    int    `json:"accural"`
		UploadedAt string `json:"uploaded_at"`
	}

	responses := make([]*response, 0, len(orders))

	for _, order := range orders {
		responses = append(responses, &response{
			Number:     order.OrderID,
			Status:     order.Status,
			Accural:    order.PointsAccumulated,
			UploadedAt: order.CreatedAt,
		})

	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

// GetBalance - получение баланса пользователя
func (h *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	// Используем middleware для авторизации
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.loyaltyService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

// Проверка на валидность номера заказа
func isValidOrderID(orderID string) bool {
	// Убираем все пробелы из строки
	orderID = strings.ReplaceAll(orderID, " ", "")

	// Если номер заказа слишком короткий или содержит нецифровые символы, сразу возвращаем false
	if len(orderID) < 1 || len(orderID) > 15 {
		return false
	}

	// Проверяем, что все символы - цифры
	for _, ch := range orderID {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	// Применяем алгоритм Луна
	sum := 0
	shouldDouble := false

	// Идем с конца строки и применяем Луну
	for i := len(orderID) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(orderID[i]))
		if err != nil {
			return false
		}

		// Если нужно удвоить цифру
		if shouldDouble {
			digit *= 2
			// Если удвоенное число больше 9, складываем его цифры
			if digit > 9 {
				digit = digit - 9
			}
		}

		// Добавляем цифру в сумму
		sum += digit

		// Чередуем удвоение
		shouldDouble = !shouldDouble
	}

	// Если сумма делится на 10, то номер корректен
	return sum%10 == 0
}

// WithdrawPoints - вывод баллов с аккаунта пользователя
func (h *UserHandler) WithdrawPoints(w http.ResponseWriter, r *http.Request) {
	// Используем middleware для авторизации
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		Points int `json:"points"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Выводим баллы
	err := h.loyaltyService.WithdrawPoints(r.Context(), userID, request.Points)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals - получение истории выводов баллов пользователя
func (h *UserHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Используем middleware для авторизации
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.loyaltyService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}
