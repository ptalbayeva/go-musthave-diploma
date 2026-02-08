package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/ptalbayeva/go-musthave-diploma/internal/middleware"
	"github.com/stretchr/testify/require"
)

func generateToken(t *testing.T, secret string, userID uuid.UUID) string {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	return tokenStr
}

func TestAuthMiddleware_Unauthorized_NoToken(t *testing.T) {
	mw := middleware.NewAuthMiddleware("secret")

	handler := mw.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_Unauthorized_InvalidToken(t *testing.T) {
	mw := middleware.NewAuthMiddleware("secret")

	handler := mw.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_Success(t *testing.T) {
	secret := "secret"
	userID := uuid.New()

	token := generateToken(t, secret, userID)

	mw := middleware.NewAuthMiddleware(secret)

	handler := mw.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.UserIDFromContext(r.Context())
		require.True(t, ok)
		require.Equal(t, userID, id)

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
