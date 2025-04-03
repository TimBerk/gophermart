package router

import (
	"TimBerk/gophermart/internal/app/handlers"
	"TimBerk/gophermart/internal/app/middlewares/auth"
	"TimBerk/gophermart/internal/app/settings/config"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter(dataStore handlers.Store, cfg *config.Config, ctx context.Context) chi.Router {
	handler := handlers.NewHandler(dataStore, cfg, ctx)

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Group(func(r chi.Router) {
		r.Post("/api/user/register", handler.Register)
		r.Post("/api/user/login", handler.Login)
	})

	router.Group(func(r chi.Router) {
		r.Use(auth.Authentication(cfg))
		r.Get("/api/user/orders/{number}", handler.GetOrder)
		r.Post("/api/user/orders", handler.CreateOrder)
		r.Get("/api/user/orders", handler.GetOrders)

		r.Get("/api/user/balance", handler.GetBalance)
		r.Post("/api/user/balance/withdraw", handler.WithdrawBalance)
		r.Get("/api/user/withdrawals", handler.GetWithdraw)
	})

	return router
}
