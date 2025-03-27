package main

import (
	"TimBerk/gophermart/internal/app/settings/config"
	"TimBerk/gophermart/internal/app/settings/logger"
	"TimBerk/gophermart/internal/app/settings/router"
	"TimBerk/gophermart/internal/app/store"
	"TimBerk/gophermart/internal/app/worker"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
	workerUpdateCtx, cancelUpdate := context.WithCancel(ctx)
	defer cancelUpdate()

	var wgBackgroud sync.WaitGroup
	wgBackgroud.Add(1)
	go worker.UpdateStateOrders(workerUpdateCtx, cfg, pgStore, &wgBackgroud)

	// Create a channel to listen for shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	router := router.InitRouter(pgStore, cfg, ctx)
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	go func() {
		defer wgBackgroud.Done()
		logger.Log.WithField("address", cfg.RunAddress).Info("Starting server")
		if errServe := server.ListenAndServe(); errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
			logger.Log.WithField("error", errServe).Fatal("Server error")
		}
	}()

	<-shutdownChan
	logger.Log.Info("Shutdown signal received")

	// Create a context with timeout for shutdown
	ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if errShutdown := server.Shutdown(ctxShutdown); err != nil {
		logger.Log.WithField("error", errShutdown).Fatal("Server shutdown error")
	}

	// Wait for all goroutines to complete
	done := make(chan struct{})
	go func() {
		wgBackgroud.Wait()
		close(done)
	}()

	// Wait for either all goroutines to finish or timeout
	select {
	case <-done:
		logger.Log.Info("All goroutines finished")
	case <-ctxShutdown.Done():
		logger.Log.Info("Timeout waiting for goroutines to finish")
	}

	logger.Log.Info("Server exited properly")
}
