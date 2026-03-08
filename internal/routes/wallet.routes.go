package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func WalletRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /balance", middleware.Protect(middleware.IsDriver(http.HandlerFunc(controllers.GetWalletBalance))))
	mux.Handle("POST /top-up", middleware.Protect(middleware.IsDriver(http.HandlerFunc(controllers.TopUpWallet))))
	mux.Handle("PUT /admin/adjust", middleware.Protect(middleware.IsAdmin(http.HandlerFunc(controllers.AdjustWallet))))

	return mux
}
