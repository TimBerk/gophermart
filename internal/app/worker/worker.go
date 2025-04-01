package worker

import (
	"TimBerk/gophermart/internal/app/client"
	model "TimBerk/gophermart/internal/app/models/order"
	"TimBerk/gophermart/internal/app/settings/config"
	"TimBerk/gophermart/internal/app/store"
	"context"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

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
	orders, err := dataStore.GetOrdersForAccrual(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": action, "error": err}).Error("failed to get list orders")
		return
	}

	logrus.WithFields(logrus.Fields{"action": action, "count": len(orders)}).Info("started work with orders")

	for _, order := range orders {
		var newStatus store.Status

		respData, errCheck := checkOrderStatus(cfg.AccrualSystemAddress, order)
		if errCheck != nil {
			logrus.WithFields(logrus.Fields{"action": action, "order": order.Number, "error": errCheck}).Error("failed to check order status")
			continue
		}
		newStatus = store.GetConstStatus(respData.Status)
		accrual := 0.0
		if newStatus == store.Processed {
			if respData.Accrual == nil {
				logrus.WithFields(logrus.Fields{"action": action, "order": order.Number}).Error("incorrect order accrual")
				continue
			}
			accrual = *respData.Accrual
		}

		if err = dataStore.UpdateOrderStatus(ctx, order.UserID, order.Number, newStatus, accrual); err != nil {
			logrus.WithFields(logrus.Fields{"action": action, "order": order.Number, "error": err}).Error("failed to update order status")
			continue
		}

		logrus.WithFields(logrus.Fields{"action": action, "order": order.Number}).Info("status updated")
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
