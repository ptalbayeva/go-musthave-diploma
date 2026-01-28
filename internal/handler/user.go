package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/service"
)

type UserHandler struct {
	authService    *service.AuthService
	loyaltyService *service.LoyaltyService
}

func NewUserHandler(authService *service.AuthService, loyaltyService *service.LoyaltyService) *UserHandler {
	return &UserHandler{
		authService:    authService,
		loyaltyService: loyaltyService,
	}
}

// Register - регистрация нового пользователя
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Регистрируем пользователя через сервис
	userID, err := h.authService.RegisterUser(r.Context(), request.Email, request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]uuid.UUID{"user_id": userID})
}

// Login - аутентификация пользователя
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Декодируем JSON в структуру запроса
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.AuthenticateUser(r.Context(), request.Email, request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// CreateOrder - создание нового заказа и начисление баллов
func (h *UserHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID  uuid.UUID `json:"user_id"`
		OrderID string    `json:"order_id"`
		Status  string    `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.loyaltyService.AddPoints(r.Context(), request.UserID, request.OrderID, 0) // Начальные баллы - 0
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetOrders - получение списка заказов пользователя
func (h *UserHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id") // Получаем user_id из query параметра
	if userID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	// Преобразуем userID в UUID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user_id format", http.StatusBadRequest)
		return
	}

	// Получаем заказы пользователя
	orders, err := h.loyaltyService.GetOrders(r.Context(), parsedUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

// GetBalance - получение баланса пользователя
func (h *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id") // Получаем user_id из query параметра
	if userID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user_id format", http.StatusBadRequest)
		return
	}

	balance, err := h.loyaltyService.GetBalance(r.Context(), parsedUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

// GetWithdrawals - получение списка выводов средств пользователя
func (h *UserHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id") // Получаем user_id из query параметра
	if userID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user_id format", http.StatusBadRequest)
		return
	}

	// Получаем выводы пользователя
	withdrawals, err := h.loyaltyService.GetWithdrawals(r.Context(), parsedUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}

// WithdrawPoints - запрос на списание баллов с накопительного счёта
func (h *UserHandler) WithdrawPoints(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID uuid.UUID `json:"user_id"`
		Points int       `json:"points"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.loyaltyService.WithdrawPoints(r.Context(), request.UserID, request.Points)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"remaining_balance": request.Points})
}
