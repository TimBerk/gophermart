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
	GetOrdersForAccrual(ctx context.Context, status store.Status) ([]order.OrderAccrual, error)
	UpdateOrderBalance(ctx context.Context, tx pgx.Tx, order string, sum int) error
	UpdateOrderStatus(ctx context.Context, order string, status store.Status) error

	GetBalance(ctx context.Context, userID int64) (balance.Balance, error)
	UpdateBalance(ctx context.Context, tx pgx.Tx, userID int64, sum int) error
	GetOrderWithdrawals(ctx context.Context, userID int64) (balance.WithdrawnList, error)
}

func NewHandler(dataStore Store, cfg *config.Config, ctx context.Context) *Handler {
	return &Handler{dataStore, cfg, ctx}
}

func initLogFields(fields logrus.Fields) *logrus.Entry {
	return logrus.WithFields(fields)
}
