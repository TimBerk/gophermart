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

func preparedOrders(ctx context.Context, cfg *config.Config, dataStore *store.PostgresStore, action string) {
	orders, err := dataStore.GetOrdersForAccrual(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": action, "error": err}).Error("failed to get list orders")
		return
	}

	logrus.WithFields(logrus.Fields{"action": action, "count": len(orders)}).Info("started work with orders")

	for _, order := range orders {
		var newStatus store.Status

		respData, err := checkOrderStatus(cfg.AccrualSystemAddress, order)
		if err != nil {
			logrus.WithFields(logrus.Fields{"action": action, "order": order.Number, "error": err}).Error("failed to check order status")
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

func UpdateStateOrders(ctx context.Context, cfg *config.Config, dataStore *store.PostgresStore, wg *sync.WaitGroup) {
	defer wg.Done()

	action := "W.UpdateStateOrders"

	for {
		preparedOrders(ctx, cfg, dataStore, action)
		time.Sleep(2 * time.Second)
	}
}

func checkOrderStatus(url string, order model.UserOrder) (*model.OrderAccrual, error) {
	accrualClient := client.NewClient(url)
	return accrualClient.GetStatus(order.Number)
}
