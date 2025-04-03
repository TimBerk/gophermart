package worker

import (
	"TimBerk/gophermart/internal/app/client"
	model "TimBerk/gophermart/internal/app/models/order"
	"TimBerk/gophermart/internal/app/settings/config"
	"TimBerk/gophermart/internal/app/store"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type APIError struct {
	Message    string
	RetryAfter time.Duration
	StatusCode int
}

func (e *APIError) Error() string {
	return e.Message
}

type OrderStore interface {
	GetOrdersForAccrual(ctx context.Context) ([]model.UserOrder, error)
	UpdateOrderStatus(ctx context.Context, userID int64, order string, status store.Status, accrual float64) error
}

func preparedOrders(
	ctx context.Context,
	cfg *config.Config,
	dataStore OrderStore,
	action string,
	checkOrderStatus func(string, model.UserOrder) (*model.OrderAccrual, error),
) {
	logFields := logrus.WithFields(logrus.Fields{"action": action})

	orders, err := dataStore.GetOrdersForAccrual(ctx)
	if err != nil {
		logFields.WithField("error", err).Error("failed to get list orders")
		return
	}

	logFields.WithField("count", len(orders)).Info("started work with orders")

	for _, order := range orders {
		var newStatus store.Status

		respData, errCheck := checkOrderStatus(cfg.AccrualSystemAddress, order)
		if errCheck != nil {
			logFields.WithFields(logrus.Fields{"order": order.Number, "error": errCheck}).Error("failed to check order status")

			if retryAfter := getRetryAfterFromError(errCheck); retryAfter > 0 {
				logFields.WithFields(logrus.Fields{
					"order":      order.Number,
					"retryAfter": retryAfter,
				}).Info("waiting due to Retry-After")
				time.Sleep(retryAfter)
			}
			continue
		}
		newStatus = store.GetConstStatus(respData.Status)
		accrual := 0.0
		if newStatus == store.Processed {
			if respData.Accrual == nil {
				logFields.WithField("order", order.Number).Error("incorrect order accrual")
				continue
			}
			accrual = *respData.Accrual
		}

		if err = dataStore.UpdateOrderStatus(ctx, order.UserID, order.Number, newStatus, accrual); err != nil {
			logFields.WithFields(logrus.Fields{"order": order.Number, "error": err}).Error("failed to update order status")
			continue
		}

		logFields.WithField("order", order.Number).Info("status updated")
	}
}

func UpdateStateOrders(ctx context.Context, cfg *config.Config, dataStore OrderStore, wg *sync.WaitGroup) {
	defer wg.Done()

	action := "W.UpdateStateOrders"

	for {
		preparedOrders(ctx, cfg, dataStore, action, CheckOrderStatus)
		time.Sleep(2 * time.Second)
	}
}

func CheckOrderStatus(url string, order model.UserOrder) (*model.OrderAccrual, error) {
	accrualClient := client.NewClient(url)
	return accrualClient.GetStatus(order.Number)
}

func getRetryAfterFromError(err error) time.Duration {
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
		return apiErr.RetryAfter
	}
	return 0
}
