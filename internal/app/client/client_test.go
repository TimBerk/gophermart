package client

import (
	model "TimBerk/gophermart/internal/app/models/order"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	baseURL := "http://test.com"

	c := NewClient(baseURL)

	assert.Equal(t, baseURL, c.url, "Client URL should match constructor argument")
}

func TestFullPathCorrect(t *testing.T) {
	baseURL := "http://test.com"
	c := NewClient(baseURL)

	path, err := c.getFullPath("/path")

	assert.NoError(t, err, "Should not return error")
	assert.Equal(t, path, c.url+"/path", "Client URL should match constructor argument")
}

func TestFullPathInCorrect(t *testing.T) {
	baseURL := "http://%test:host"
	c := NewClient(baseURL)

	path, err := c.getFullPath("/path")

	assert.Error(t, err, "Should return error")
	assert.Equal(t, path, "", "Client URL should match constructor argument")
}

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name           string
		orderID        string
		mockResponse   interface{}
		mockStatus     int
		wantResult     *model.OrderAccrual
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:    "successful response",
			orderID: "50405077004",
			mockResponse: map[string]interface{}{
				"order":   "50405077004",
				"status":  "PROCESSED",
				"accrual": 100.50,
			},
			mockStatus: http.StatusOK,
			wantResult: &model.OrderAccrual{
				Number:  "50405077004",
				Status:  "PROCESSED",
				Accrual: func() *float64 { f := 100.50; return &f }(),
			},
			wantErr: false,
		},
		{
			name:           "order not ready",
			orderID:        "99999",
			mockResponse:   "",
			mockStatus:     http.StatusNoContent,
			wantResult:     nil,
			wantErr:        true,
			expectedErrMsg: "order not ready",
		},
		{
			name:           "too many requests",
			orderID:        "50405077004",
			mockResponse:   "",
			mockStatus:     http.StatusTooManyRequests,
			wantResult:     nil,
			wantErr:        true,
			expectedErrMsg: "order not ready",
		},
		{
			name:    "invalid json response",
			orderID: "50405077004",
			mockResponse: `{
				"order": "50405077004",
				"status": "PROCESSED",
				"accrual": "should be number"
			}`,
			mockStatus:     http.StatusOK,
			wantResult:     nil,
			wantErr:        true,
			expectedErrMsg: "expected number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/orders/"+tt.orderID, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tt.mockStatus)

				switch v := tt.mockResponse.(type) {
				case string:
					w.Write([]byte(v))
				default:
					json.NewEncoder(w).Encode(v)
				}
			}))
			defer ts.Close()
			c := NewClient(ts.URL)

			got, err := c.GetStatus(tt.orderID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, got)
			}
		})
	}
}
