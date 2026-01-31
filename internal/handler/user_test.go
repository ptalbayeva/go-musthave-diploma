package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/handler"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	RegisterFn     func(ctx context.Context, email, password string) (uuid.UUID, error)
	AuthenticateFn func(ctx context.Context, email, password string) (string, error)
}

func (m mockAuthService) RegisterUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	return m.RegisterFn(ctx, email, password)
}

func (m mockAuthService) AuthenticateUser(ctx context.Context, email, password string) (string, error) {
	return m.AuthenticateFn(ctx, email, password)
}

func TestUserHandler_Login_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		authFn         func(ctx context.Context, email, password string) (string, error)
		expectedStatus int
		expectedToken  string
	}{
		{
			name: "success login",
			requestBody: map[string]string{
				"login":    "test@mail.com",
				"password": "123456",
			},
			authFn: func(ctx context.Context, email, password string) (string, error) {
				return "jwt-token", nil
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "jwt-token",
		},
		{
			name:           "invalid json",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "auth error",
			requestBody: map[string]string{
				"login":    "test@mail.com",
				"password": "wrong",
			},
			authFn: func(ctx context.Context, email, password string) (string, error) {
				return "", errors.New("unauthorized")
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			authService := &mockAuthService{
				AuthenticateFn: tt.authFn,
			}
			h := handler.NewUserHandler(authService, nil, "secret", nil)

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]string
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tt.expectedToken, resp["token"])

				cookies := rec.Result().Cookies()
				require.Len(t, cookies, 1)
				require.Equal(t, "auth_token", cookies[0].Name)
			}
		})
	}
}
