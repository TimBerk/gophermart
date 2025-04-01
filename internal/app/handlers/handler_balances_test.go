package handlers

import (
	"TimBerk/gophermart/internal/app/middlewares/auth"
	model "TimBerk/gophermart/internal/app/models/balance"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetBalance(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	tests := []struct {
		name         string
		setupMocks   func(*MockStore)
		setupRequest func() *http.Request
		expectedCode int
		expectedBody string
	}{
		{
			name: "successful balance retrieval",
			setupMocks: func(store *MockStore) {
				store.On("GetBalance", mock.Anything, mockUserID).Return(
					model.Balance{
						Current:   100.5,
						Withdrawn: 20.0,
					}, nil)
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/balance", nil)
				reqCtx := context.WithValue(req.Context(), auth.UserIDKey, mockUserID)
				return req.WithContext(reqCtx)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"current":100.5,"withdrawn":20}`,
		},
		{
			name:       "unauthorized access",
			setupMocks: func(store *MockStore) {},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/balance", nil)
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"User is not authorized"}`,
		},
		{
			name: "database error",
			setupMocks: func(store *MockStore) {
				store.On("GetBalance", mock.Anything, mockUserID).Return(
					model.Balance{}, errors.New("db error"))
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/balance", nil)
				reqCtx := context.WithValue(req.Context(), auth.UserIDKey, mockUserID)
				return req.WithContext(reqCtx)
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"failed to find balance"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockStore)
			tt.setupMocks(mockStore)
			handler := &Handler{store: mockStore, ctx: context.Background()}

			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			handler.GetBalance(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			mockStore.AssertExpectations(t)
		})
	}
}

func TestWithdrawBalance(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	tests := []struct {
		name           string
		setupMocks     func(*MockStore)
		requestBody    interface{}
		isAuth         bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful withdrawal",
			setupMocks: func(store *MockStore) {
				store.On("GetBalance", mock.Anything, mockUserID).Return(
					model.Balance{Current: 100.0}, nil)
				store.On("AddWithdrawal", mock.Anything, mockUserID, mockOrderID, 50.0).Return(nil)
			},
			requestBody:    model.WithdrawnRequest{Number: mockOrderID, Sum: 50.0},
			isAuth:         true,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "unauthorized access",
			setupMocks:     func(*MockStore) {},
			requestBody:    model.WithdrawnRequest{Number: mockOrderID, Sum: 50.0},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"User is not authorized"}`,
		},
		{
			name: "empty balance",
			setupMocks: func(store *MockStore) {
				store.On("GetBalance", mock.Anything, mockUserID).Return(
					model.Balance{Current: 0.0}, nil)
			},
			requestBody:    model.WithdrawnRequest{Number: mockOrderID, Sum: 50.0},
			isAuth:         true,
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   `{"error":"failed to use balance: it's empty"}`,
		},
		{
			name: "invalid request body",
			setupMocks: func(store *MockStore) {
				store.On("GetBalance", mock.Anything, mockUserID).Return(
					model.Balance{Current: 100.0}, nil)
			},
			requestBody:    "invalid json",
			isAuth:         true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to parse request data"}`,
		},
		{
			name: "validation error",
			setupMocks: func(store *MockStore) {
				store.On("GetBalance", mock.Anything, mockUserID).Return(
					model.Balance{Current: 100.0}, nil)
			},
			requestBody:    model.WithdrawnRequest{Number: "", Sum: 50.0},
			isAuth:         true,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"failed to validate request data"}`,
		},
		{
			name: "insufficient funds",
			setupMocks: func(store *MockStore) {
				store.On("GetBalance", mock.Anything, mockUserID).Return(
					model.Balance{Current: 30.0}, nil)
			},
			requestBody:    model.WithdrawnRequest{Number: mockOrderID, Sum: 50.0},
			isAuth:         true,
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   `{"error":"failed to use balance: it's less than sum"}`,
		},
		{
			name: "database error on withdrawal",
			setupMocks: func(store *MockStore) {
				store.On("GetBalance", mock.Anything, mockUserID).Return(
					model.Balance{Current: 100.0}, nil)
				store.On("AddWithdrawal", mock.Anything, mockUserID, mockOrderID, 50.0).Return(
					errors.New("db error"))
			},
			requestBody:    model.WithdrawnRequest{Number: mockOrderID, Sum: 50.0},
			isAuth:         true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to update order"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockStore)
			tt.setupMocks(mockStore)
			h := &Handler{store: mockStore, ctx: context.Background()}

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.isAuth {
				reqCtx := context.WithValue(req.Context(), auth.UserIDKey, mockUserID)
				req = req.WithContext(reqCtx)
			}
			rr := httptest.NewRecorder()

			h.WithdrawBalance(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestHandler_GetWithdraw(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)
	testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		setupMocks     func(*MockStore)
		isAuth         bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful get withdrawals",
			setupMocks: func(store *MockStore) {
				store.On("GetOrderWithdrawals", mock.Anything, mockUserID).Return(
					model.WithdrawnList{
						model.WithdrawnResponse{Number: mockOrderID, Sum: 50.0, CreatedAt: testTime},
					}, nil)
			},
			isAuth:         true,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"order":"50405077004","sum":50,"processed_at":"2023-01-01T00:00:00Z"}]`,
		},
		{
			name:           "unauthorized access",
			setupMocks:     func(*MockStore) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"User is not authorized"}`,
		},
		{
			name: "no withdrawals found",
			setupMocks: func(store *MockStore) {
				store.On("GetOrderWithdrawals", mock.Anything, mockUserID).Return(
					model.WithdrawnList{}, nil)
			},
			isAuth:         true,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name: "no rows error - treated as empty list",
			setupMocks: func(store *MockStore) {
				store.On("GetOrderWithdrawals", mock.Anything, mockUserID).Return(
					model.WithdrawnList{}, pgx.ErrNoRows)
			},
			isAuth:         true,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name: "database error",
			setupMocks: func(store *MockStore) {
				store.On("GetOrderWithdrawals", mock.Anything, mockUserID).Return(
					model.WithdrawnList{}, errors.New("db error"))
			},
			isAuth:         true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to find orders"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockStore)
			tt.setupMocks(mockStore)
			h := &Handler{store: mockStore, ctx: context.Background()}

			req := httptest.NewRequest("GET", "/withdrawals", nil)
			if tt.isAuth {
				reqCtx := context.WithValue(req.Context(), auth.UserIDKey, mockUserID)
				req = req.WithContext(reqCtx)
			}
			rr := httptest.NewRecorder()

			h.GetWithdraw(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else if tt.expectedStatus == http.StatusNoContent {
				assert.Empty(t, rr.Body.String())
			}
			mockStore.AssertExpectations(t)
		})
	}
}
