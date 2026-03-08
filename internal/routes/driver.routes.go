package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func DriverRoutes() http.Handler {
	mux := http.NewServeMux()

	// Authenticated users (customers) finding drivers
	mux.Handle("GET /nearby", middleware.Protect(http.HandlerFunc(controllers.GetNearbyDrivers)))

	// Admin-only routes
	mux.Handle("POST /{id}/block", middleware.Protect(middleware.IsAdmin(http.HandlerFunc(controllers.BlockDriver))))
	mux.Handle("POST /{id}/unblock", middleware.Protect(middleware.IsAdmin(http.HandlerFunc(controllers.UnblockDriver))))

	// Driver-specific routes
	mux.Handle("GET /rides", middleware.Protect(middleware.IsDriver(http.HandlerFunc(controllers.GetDriverRides))))

	return mux
}
