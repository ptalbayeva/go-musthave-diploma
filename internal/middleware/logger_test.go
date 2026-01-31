package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ptalbayeva/go-musthave-diploma/internal/middleware"
	"github.com/stretchr/testify/require"
)

// TestRequestLogger проверка логирования запросов
func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   string
		expectedSize   int
	}{
		{
			name: "200 ok with body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("hello"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "hello",
			expectedSize:   len("hello"),
		},
		{
			name: "204 no content",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
			expectedSize:   0,
		},
		{
			name: "implicit 200 without WriteHeader",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("data"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "data",
			expectedSize:   len("data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware.Log = middleware.Log.WithOptions()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			loggerMw := middleware.RequestLogger()
			handler := loggerMw(tt.handler)

			handler.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			body := rec.Body.String()

			require.Equal(t, tt.expectedStatus, res.StatusCode)
			require.Equal(t, tt.expectedBody, body)
			require.Equal(t, tt.expectedSize, len(body))
		})
	}
}
