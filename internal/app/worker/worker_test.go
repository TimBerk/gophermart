package worker

import (
	model "TimBerk/gophermart/internal/app/models/order"
	"TimBerk/gophermart/internal/app/settings/config"
	handlerStore "TimBerk/gophermart/internal/app/store"
	"TimBerk/gophermart/pkg/utils"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) GetOrdersForAccrual(ctx context.Context) ([]model.UserOrder, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]model.UserOrder), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStore) UpdateOrderStatus(ctx context.Context, userID int64, order string, status handlerStore.Status, accrual float64) error {
	args := m.Called(ctx, userID, order, status, accrual)
	return args.Error(0)
}

func TestPreparedOrders(t *testing.T) {
	mockCtx := context.Background()
	mockCfg := &config.Config{AccrualSystemAddress: "http://test"}

	tests := []struct {
		name          string
		mockSetup     func(*MockStore) func(string, model.UserOrder) (*model.OrderAccrual, error)
		expectError   bool
		expectUpdates int
	}{
		{
			name: "successful order processing",
			mockSetup: func(store *MockStore) func(string, model.UserOrder) (*model.OrderAccrual, error) {
				order := model.UserOrder{
					Number: "123456",
					UserID: int64(777),
				}
				store.On("GetOrdersForAccrual", mock.Anything).Return([]model.UserOrder{order}, nil)
				store.On("UpdateOrderStatus",
					mock.Anything,
					int64(777),
					"123456",
					handlerStore.Processed,
					100.0).Return(nil)

				return func(addr string, o model.UserOrder) (*model.OrderAccrual, error) {
					return &model.OrderAccrual{
						Status:  "PROCESSED",
						Accrual: utils.PtrFloat64(100.0),
					}, nil
				}
			},
			expectUpdates: 1,
		},
		{
			name: "error getting orders",
			mockSetup: func(store *MockStore) func(string, model.UserOrder) (*model.OrderAccrual, error) {
				store.On("GetOrdersForAccrual", mock.Anything).Return(nil, errors.New("db error"))
				return nil
			},
			expectError: true,
		},
		{
			name: "error checking order status",
			mockSetup: func(store *MockStore) func(string, model.UserOrder) (*model.OrderAccrual, error) {
				order := model.UserOrder{Number: "123456", UserID: 777}
				store.On("GetOrdersForAccrual", mock.Anything).Return([]model.UserOrder{order}, nil)

				return func(addr string, o model.UserOrder) (*model.OrderAccrual, error) {
					return nil, errors.New("network error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockStore)
			checker := tt.mockSetup(mockStore)

			preparedOrders(mockCtx, mockCfg, mockStore, "test", checker)

			if tt.expectUpdates > 0 {
				mockStore.AssertNumberOfCalls(t, "UpdateOrderStatus", tt.expectUpdates)
			}
			mockStore.AssertExpectations(t)
		})
	}
}
