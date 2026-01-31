package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ptalbayeva/go-musthave-diploma/internal/client"
	"github.com/stretchr/testify/require"
)

func TestClient_GetOrder_TableDriven(t *testing.T) {
	accrual := 123.45

	tests := []struct {
		name           string
		statusCode     int
		responseBody   any
		retryAfter     string
		expectedStatus int
		expectedRetry  int
		expectError    bool
	}{
		{
			name:       "200 ok",
			statusCode: http.StatusOK,
			responseBody: client.OrderResponse{
				Order:   "123",
				Status:  "PROCESSED",
				Accrual: &accrual,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "204 no content",
			statusCode:     http.StatusNoContent,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "429 too many requests default retry",
			statusCode:     http.StatusTooManyRequests,
			expectedStatus: http.StatusTooManyRequests,
			expectedRetry:  60,
		},
		{
			name:           "500 internal server error",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "unexpected status",
			statusCode:     http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// fake accrual server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)

				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}

				if tt.responseBody != nil {
					_ = json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			c := client.NewClient(server.URL)

			res, status, retryAfter, err := c.GetOrder("123")

			require.Equal(t, tt.expectedStatus, status)
			require.Equal(t, tt.expectedRetry, retryAfter)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, res)
				return
			}

			require.NoError(t, err)

			if tt.statusCode == http.StatusOK {
				require.NotNil(t, res)
				require.Equal(t, "123", res.Order)
			} else {
				require.Nil(t, res)
			}
		})
	}
}
