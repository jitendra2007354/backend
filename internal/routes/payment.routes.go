package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func PaymentRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /make-payment", middleware.Protect(http.HandlerFunc(controllers.MakePayment)))

	return mux
}
