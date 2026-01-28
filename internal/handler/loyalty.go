package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/service"
)

// Хендлер для работы с лояльностью
type LoyaltyHandler struct {
	loyaltyService *service.LoyaltyService
}

func NewLoyaltyHandler(loyaltyService *service.LoyaltyService) *LoyaltyHandler {
	return &LoyaltyHandler{loyaltyService: loyaltyService}
}

// AddPoints обработка начисления баллов
func (h *LoyaltyHandler) AddPoints(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID  uuid.UUID `json:"user_id"`
		OrderID string    `json:"order_id"`
		Points  int       `json:"points"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Начисление баллов
	if err := h.loyaltyService.AddPoints(r.Context(), request.UserID, request.OrderID, request.Points); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "points added"})
}

// WithdrawPoints обработка списания баллов
func (h *LoyaltyHandler) WithdrawPoints(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID uuid.UUID `json:"user_id"`
		Points int       `json:"points"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.loyaltyService.WithdrawPoints(r.Context(), request.UserID, request.Points); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "points withdrawn"})
}

// GetBalance получить баланс пользователя
func (h *LoyaltyHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	balance, err := h.loyaltyService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

// GetWithdrawals получить историю выводов баллов пользователя
func (h *LoyaltyHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	withdrawals, err := h.loyaltyService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(withdrawals)
}

// CreateOrder обработка загрузки номера заказа
func (h *LoyaltyHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID  uuid.UUID `json:"user_id"`
		OrderID string    `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаём заказ (необходимо добавить в репозиторий)
	err := h.loyaltyService.AddPoints(r.Context(), request.UserID, request.OrderID, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
