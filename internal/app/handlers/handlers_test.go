package handlers

import (
	"TimBerk/gophermart/internal/app/models/balance"
	"TimBerk/gophermart/internal/app/models/order"
	"TimBerk/gophermart/internal/app/store"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"time"
)

var (
	mockUserID  = int64(777)
	mockOrderID = "50405077004"
	mockTime    = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) BeginTx(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockStore) AddUser(ctx context.Context, username string, password string) (int64, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) CheckUser(ctx context.Context, username string) (int64, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) GetUser(ctx context.Context, username string) (store.UserRecord, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(store.UserRecord), args.Error(1)
}

func (m *MockStore) AddOrder(ctx context.Context, userID int64, order string) error {
	args := m.Called(ctx, userID, order)
	return args.Error(0)
}

func (m *MockStore) GetOrder(ctx context.Context, order string) (store.OrderRecord, error) {
	args := m.Called(ctx, order)
	return args.Get(0).(store.OrderRecord), args.Error(1)
}

func (m *MockStore) GetOrderList(ctx context.Context, userID int64) (order.OrderListResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(order.OrderListResponse), args.Error(1)
}

func (m *MockStore) GetOrdersForAccrual(ctx context.Context) ([]order.UserOrder, error) {
	args := m.Called(ctx)
	return args.Get(0).([]order.UserOrder), args.Error(1)
}

func (m *MockStore) AddWithdrawal(ctx context.Context, userID int64, order string, sum float64) error {
	args := m.Called(ctx, userID, order, sum)
	return args.Error(0)
}

func (m *MockStore) UpdateOrderStatus(ctx context.Context, userID int64, order string, status store.Status, accrual float64) error {
	args := m.Called(ctx, userID, order, status, accrual)
	return args.Error(0)
}

func (m *MockStore) GetBalance(ctx context.Context, userID int64) (balance.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(balance.Balance), args.Error(1)
}

func (m *MockStore) AddBalance(ctx context.Context, tx pgx.Tx, userID int64, sum float64) error {
	args := m.Called(ctx, tx, userID, sum)
	return args.Error(0)
}

func (m *MockStore) WithdrawBalance(ctx context.Context, tx pgx.Tx, userID int64, sum float64) error {
	args := m.Called(ctx, tx, userID, sum)
	return args.Error(0)
}

func (m *MockStore) GetOrderWithdrawals(ctx context.Context, userID int64) (balance.WithdrawnList, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(balance.WithdrawnList), args.Error(1)
}
