package main

import (
	"TimBerk/gophermart/internal/app/settings/config"
	"TimBerk/gophermart/internal/app/settings/logger"
	"TimBerk/gophermart/internal/app/settings/router"
	"TimBerk/gophermart/internal/app/store"
	"TimBerk/gophermart/internal/app/worker"
	"context"
	"net/http"
	"sync"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()
	logger.Initialize(cfg.LogLevel)

	pgStore, err := store.NewPostgresStore(cfg)
	if err != nil {
		logger.Log.Fatal("Read Store: ", err)
	}

	// Workers for updating order status
	workerNewCtx, cancelNew := context.WithCancel(ctx)
	defer cancelNew()
	workerUpdateCtx, cancelUpdate := context.WithCancel(ctx)
	defer cancelUpdate()

	var wgNew sync.WaitGroup
	wgNew.Add(1)
	go worker.RegisterOrders(workerNewCtx, cfg, pgStore, &wgNew)

	var wgCheck sync.WaitGroup
	wgCheck.Add(1)
	go worker.UpdateStateOrders(workerUpdateCtx, cfg, pgStore, &wgCheck)

	router := router.InitRouter(pgStore, cfg, ctx)
	logger.Log.WithField("address", cfg.RunAddress).Info("Starting server")
	err = http.ListenAndServe(cfg.RunAddress, router)
	if err != nil {
		logger.Log.Fatal("ListenAndServe: ", err)
	}

	wgNew.Wait()
	wgCheck.Wait()
}
