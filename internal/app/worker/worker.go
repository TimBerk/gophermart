package worker

import (
	"TimBerk/gophermart/internal/app/client"
	model "TimBerk/gophermart/internal/app/models/order"
	"TimBerk/gophermart/internal/app/settings/config"
	"TimBerk/gophermart/internal/app/store"
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

func preparedOrders(ctx context.Context, cfg *config.Config, dataStore *store.PostgresStore, status store.Status, action string) {
	orders, err := dataStore.GetOrdersForAccrual(ctx, status)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": action, "error": err}).Error("failed to get list orders")
		return
	}

	logrus.WithFields(logrus.Fields{"action": action, "count": len(orders)}).Info("started work with orders")

	for _, order := range orders {
		var newStatus store.Status

		if status == store.New {
			respStatus, err := registerNewOrder(cfg.AccrualSystemAddress, order)
			if err != nil || respStatus != http.StatusAccepted {
				logrus.WithFields(logrus.Fields{"action": action, "order": order.Number, "error": err}).Error("failed to register order status")
				continue
			}

			newStatus = store.Processing
		} else {
			respStatus, err := checkOrderStatus(cfg.AccrualSystemAddress, order)
			if err != nil {
				logrus.WithFields(logrus.Fields{"action": action, "order": order.Number, "error": err}).Error("failed to check order status")
				continue
			}

			newStatus = store.GetConstStatus(respStatus)
		}

		if err = dataStore.UpdateOrderStatus(ctx, order.Number, newStatus); err != nil {
			logrus.WithFields(logrus.Fields{"action": action, "order": order.Number, "error": err}).Error("failed to update order status")
			continue
		}

		logrus.WithFields(logrus.Fields{"action": action, "order": order.Number}).Info("status updated")
	}

	time.Sleep(2 * time.Second)
}

func RegisterOrders(ctx context.Context, cfg *config.Config, dataStore *store.PostgresStore, wg *sync.WaitGroup) {
	defer wg.Done()

	action := "W.RegisterOrders"

	for {
		preparedOrders(ctx, cfg, dataStore, store.New, action)
	}
}

func UpdateStateOrders(ctx context.Context, cfg *config.Config, dataStore *store.PostgresStore, wg *sync.WaitGroup) {
	defer wg.Done()

	action := "W.UpdateStateOrders"

	for {
		preparedOrders(ctx, cfg, dataStore, store.Processing, action)
	}
}

func registerNewOrder(url string, order model.OrderAccrual) (int, error) {
	accrualClient := client.NewClient(url)
	return accrualClient.Register(order.Number)
}

func checkOrderStatus(url string, order model.OrderAccrual) (string, error) {
	accrualClient := client.NewClient(url)
	return accrualClient.GetStatus(order.Number)
}
