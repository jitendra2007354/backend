package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func CustomerRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /rides", middleware.Protect(middleware.IsCustomer(http.HandlerFunc(controllers.GetCustomerRides))))
	
	return mux
}
