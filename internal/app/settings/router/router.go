package router

import (
	"TimBerk/gophermart/internal/app/handlers"
	"TimBerk/gophermart/internal/app/middlewares/auth"
	"TimBerk/gophermart/internal/app/middlewares/rate_limit"
	"TimBerk/gophermart/internal/app/settings/config"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter(dataStore handlers.Store, cfg *config.Config, ctx context.Context) chi.Router {
	rateLimiter := rate_limit.NewRateLimiter()
	handler := handlers.NewHandler(dataStore, cfg, ctx)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(auth.Authentication(cfg))
	router.Use(rate_limit.RateLimit(rateLimiter))

	router.Post("/api/user/register", handler.Register)
	router.Post("/api/user/login", handler.Login)

	router.Get("/api/user/orders/{number}", handler.GetOrder)
	router.Post("/api/user/orders", handler.CreateOrder)
	router.Get("/api/user/orders", handler.GetOrders)

	router.Get("/api/user/balance", handler.GetBalance)
	router.Post("/api/user/balance/withdraw", handler.WithdrawBalance)
	router.Get("/api/user/withdrawals", handler.GetWithdraw)

	return router
}
