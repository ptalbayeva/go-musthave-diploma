package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	SecretKey string
}

func NewAuthMiddleware(secretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		SecretKey: secretKey,
	}
}

func (m *AuthMiddleware) Authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Попытка извлечь токен из cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil || cookie.Value == "" {
			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Fields(authorizationHeader)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			cookie = &http.Cookie{Value: parts[1]}
		}

		// Извлечение и валидация JWT токена
		tokenStr := cookie.Value
		claims := &jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.SecretKey), nil
		})

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Извлекаем user_id из токена
		userIDStr, ok := (*claims)["user_id"].(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Парсим user_id в UUID
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Добавляем user_id в контекст запроса
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", userID)

		// Передаем управление дальше
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
