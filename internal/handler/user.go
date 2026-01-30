package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/client"
	"github.com/ptalbayeva/go-musthave-diploma/internal/middleware"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
	"github.com/ptalbayeva/go-musthave-diploma/internal/service"
	"github.com/ptalbayeva/go-musthave-diploma/internal/worker"
	"github.com/theplant/luhn"
)

type UserHandler struct {
	authService    *service.AuthService
	loyaltyService *service.LoyaltyService
	secretKey      string
	authMiddleware *middleware.AuthMiddleware
	accrualClient  *client.Client
}

func NewUserHandler(authService *service.AuthService, loyaltyService *service.LoyaltyService, secretKey string, accrualClient *client.Client) *UserHandler {
	authMiddleware := middleware.NewAuthMiddleware(secretKey)
	return &UserHandler{
		authService:    authService,
		loyaltyService: loyaltyService,
		secretKey:      secretKey,
		authMiddleware: authMiddleware,
		accrualClient:  accrualClient,
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
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	order := strings.TrimSpace(string(body))
	orderID, err := strconv.Atoi(order)
	if err != nil {
		http.Error(w, "invalid order format", http.StatusBadRequest)
		return
	}

	if !luhn.Valid(orderID) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}

	exists, ownerID, err := h.loyaltyService.OrderExists(ctx, order)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if exists {
		if ownerID == userID {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "order belongs to another user", http.StatusConflict)
		return
	}

	if err := h.loyaltyService.CreateOrder(ctx, userID, order); err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	go worker.AwaitOrderProcessed(
		context.Background(),
		order,
		userID,
		h.accrualClient,
		h.loyaltyService,
	)

	w.WriteHeader(http.StatusAccepted)
}

// GetOrders - получение списка заказов пользователя
func (h *UserHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.loyaltyService.GetOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	type response struct {
		Number     string   `json:"number"`
		Status     string   `json:"status"`
		Accrual    *float32 `json:"accrual,omitempty"`
		UploadedAt string   `json:"uploaded_at"`
	}

	responses := make([]*response, 0, len(orders))
	for _, order := range orders {
		var accrual *float32
		if order.PointsAccumulated > 0 {
			accrual = &order.PointsAccumulated
		}

		responses = append(responses, &response{
			Number:     order.OrderID,
			Status:     order.Status,
			Accrual:    accrual,
			UploadedAt: order.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

// GetBalance - получение баланса пользователя
func (h *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, withdrawn, err := h.loyaltyService.GetUserBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	resp := map[string]interface{}{
		"current":   balance,
		"withdrawn": withdrawn,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// WithdrawPoints - вывод баллов с аккаунта пользователя
func (h *UserHandler) WithdrawPoints(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orderNum, err := strconv.Atoi(req.Order)
	if err != nil || !luhn.Valid(orderNum) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	err = h.loyaltyService.WithdrawPoints(
		r.Context(),
		userID,
		req.Order,
		req.Sum,
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientBalance):
			http.Error(w, "Insufficient balance", http.StatusPaymentRequired)
			return
		case errors.Is(err, service.ErrOrderAlreadyUsed):
			http.Error(w, "Order already used", http.StatusConflict)
			return
		default:
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals - получение истории выводов баллов пользователя
func (h *UserHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	transactions, err := h.loyaltyService.GetWithdrawals(ctx, userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(transactions) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]map[string]interface{}, 0, len(transactions))
	for _, t := range transactions {
		resp = append(resp, map[string]interface{}{
			"order":        t.OrderID,
			"sum":          t.Points,
			"processed_at": t.ProcessedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
