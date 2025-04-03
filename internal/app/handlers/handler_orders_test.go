package handlers

import (
	"TimBerk/gophermart/internal/app/middlewares/auth"
	"TimBerk/gophermart/internal/app/models/order"
	storeModel "TimBerk/gophermart/internal/app/store"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

var accrual = 100.0

func TestCreateOrder(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	tests := []struct {
		name           string
		userID         int64
		requestBody    string
		setupMocks     func(*MockStore)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful order creation",
			requestBody: "50405077004", // Валидный номер заказа
			setupMocks: func(store *MockStore) {
				store.On("GetOrder", mock.Anything, "50405077004").Return(
					storeModel.OrderRecord{}, pgx.ErrNoRows)
				store.On("AddOrder", mock.Anything, mockUserID, "50405077004").Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "invalid order number",
			requestBody:    "invalid",
			setupMocks:     func(*MockStore) {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"failed to validate order number"}`,
		},
		{
			name:        "order already exists for this user",
			requestBody: "50405077004",
			setupMocks: func(store *MockStore) {
				store.On("GetOrder", mock.Anything, "50405077004").Return(
					storeModel.OrderRecord{UserID: mockUserID}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "order exists for another user",
			requestBody: "50405077004",
			setupMocks: func(store *MockStore) {
				store.On("GetOrder", mock.Anything, "50405077004").Return(
					storeModel.OrderRecord{UserID: int64(888)}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"failed to check order: it was uploaded another user"}`,
		},
		{
			name:        "database error when getting order",
			requestBody: "50405077004",
			setupMocks: func(store *MockStore) {
				store.On("GetOrder", mock.Anything, "50405077004").Return(
					storeModel.OrderRecord{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to find order"}`,
		},
		{
			name:        "database error when adding order",
			requestBody: "50405077004",
			setupMocks: func(store *MockStore) {
				store.On("GetOrder", mock.Anything, "50405077004").Return(
					storeModel.OrderRecord{}, pgx.ErrNoRows)
				store.On("AddOrder", mock.Anything, mockUserID, "50405077004").Return(
					errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to create order"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockStore)
			tt.setupMocks(mockStore)
			h := &Handler{store: mockStore, ctx: context.Background()}

			req := httptest.NewRequest("POST", "/orders", bytes.NewBufferString(tt.requestBody))
			ctx := context.WithValue(req.Context(), auth.UserIDKey, mockUserID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			h.CreateOrder(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Empty(t, rr.Body.String())
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetOrders(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)

	tests := []struct {
		name           string
		setupMocks     func(*MockStore)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success with orders",
			setupMocks: func(store *MockStore) {
				store.On("GetOrderList", mock.Anything, mockUserID).Return(
					order.OrderListResponse{
						order.OrderResponse{Number: "123", Status: "PROCESSED", Accrual: &accrual, CreatedAt: mockTime},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"number":"123","status":"PROCESSED","accrual":100,"uploaded_at":"2023-01-01T00:00:00Z"}]`,
		},
		{
			name: "no orders found",
			setupMocks: func(store *MockStore) {
				store.On("GetOrderList", mock.Anything, mockUserID).Return(
					order.OrderListResponse{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "database error",
			setupMocks: func(store *MockStore) {
				store.On("GetOrderList", mock.Anything, mockUserID).Return(
					order.OrderListResponse{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to find orders"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockStore)
			tt.setupMocks(mockStore)
			h := &Handler{store: mockStore, ctx: context.Background()}

			req := httptest.NewRequest("GET", "/orders", nil)
			ctx := context.WithValue(req.Context(), auth.UserIDKey, mockUserID)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			h.GetOrders(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Empty(t, rr.Body.String())
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetOrder(t *testing.T) {
	logrus.SetLevel(logrus.PanicLevel)
	validOrder := storeModel.OrderRecord{
		Order:   mockOrderID,
		UserID:  mockUserID,
		Status:  "PROCESSED",
		Accrual: sql.NullFloat64{Float64: accrual, Valid: true},
	}

	tests := []struct {
		name           string
		orderNumber    string
		setupMocks     func(*MockStore)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful get order",
			orderNumber: mockOrderID,
			setupMocks: func(store *MockStore) {
				store.On("GetOrder", mock.Anything, mockOrderID).Return(validOrder, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"order":"50405077004","status":"PROCESSED","accrual":100}`,
		},
		{
			name:           "invalid order number",
			orderNumber:    "invalid",
			setupMocks:     func(*MockStore) {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"failed to validate order number"}`,
		},
		{
			name:        "order not found",
			orderNumber: mockOrderID,
			setupMocks: func(store *MockStore) {
				store.On("GetOrder", mock.Anything, mockOrderID).Return(storeModel.OrderRecord{}, pgx.ErrNoRows)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "database error",
			orderNumber: mockOrderID,
			setupMocks: func(store *MockStore) {
				store.On("GetOrder", mock.Anything, mockOrderID).Return(storeModel.OrderRecord{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to find order"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockStore)
			tt.setupMocks(mockStore)
			h := &Handler{store: mockStore, ctx: context.Background()}

			req := httptest.NewRequest("GET", "/orders/"+tt.orderNumber, nil)
			ctx := context.WithValue(req.Context(), auth.UserIDKey, mockUserID)
			req = req.WithContext(ctx)

			// Устанавливаем параметр маршрута
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("number", tt.orderNumber)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.GetOrder(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Empty(t, rr.Body.String())
			}
			mockStore.AssertExpectations(t)
		})
	}
}
