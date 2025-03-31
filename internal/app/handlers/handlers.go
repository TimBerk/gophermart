package handlers

import (
	"TimBerk/gophermart/internal/app/models/balance"
	"TimBerk/gophermart/internal/app/models/order"
	"TimBerk/gophermart/internal/app/settings/config"
	"TimBerk/gophermart/internal/app/store"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	store Store
	cfg   *config.Config
	ctx   context.Context
}

type Store interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)

	AddUser(ctx context.Context, username string, password string) (int64, error)
	CheckUser(ctx context.Context, username string) (int64, error)
	GetUser(ctx context.Context, username string) (store.UserRecord, error)

	AddOrder(ctx context.Context, userID int64, order string) error
	GetOrder(ctx context.Context, order string) (store.OrderRecord, error)
	GetOrderList(ctx context.Context, userID int64) (order.OrderListResponse, error)
	GetOrdersForAccrual(ctx context.Context) ([]order.UserOrder, error)
	AddWithdrawal(ctx context.Context, userID int64, order string, sum float64) error
	UpdateOrderStatus(ctx context.Context, userID int64, order string, status store.Status, accrual float64) error

	GetBalance(ctx context.Context, userID int64) (balance.Balance, error)
	AddBalance(ctx context.Context, tx pgx.Tx, userID int64, sum float64) error
	WithdrawBalance(ctx context.Context, tx pgx.Tx, userID int64, sum float64) error
	GetOrderWithdrawals(ctx context.Context, userID int64) (balance.WithdrawnList, error)
}

func NewHandler(dataStore Store, cfg *config.Config, ctx context.Context) *Handler {
	return &Handler{dataStore, cfg, ctx}
}

func initLogFields(fields logrus.Fields) *logrus.Entry {
	return logrus.WithFields(fields)
}
