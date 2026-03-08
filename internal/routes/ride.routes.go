package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func RideRoutes() http.Handler {
	mux := http.NewServeMux()

	// All ride routes are protected
	mux.Handle("POST /", middleware.Protect(http.HandlerFunc(controllers.CreateRide)))
	mux.Handle("GET /available", middleware.Protect(http.HandlerFunc(controllers.GetAvailableRides)))
	mux.Handle("GET /{id}", middleware.Protect(http.HandlerFunc(controllers.GetRideByID)))
	mux.Handle("POST /{id}/accept", middleware.Protect(http.HandlerFunc(controllers.AcceptRide)))
	mux.Handle("POST /{id}/reject", middleware.Protect(http.HandlerFunc(controllers.RejectRide)))
	mux.Handle("POST /{id}/cancel", middleware.Protect(http.HandlerFunc(controllers.CancelRide)))
	mux.Handle("POST /{rideId}/confirm", middleware.Protect(http.HandlerFunc(controllers.ConfirmRide)))
	mux.Handle("PATCH /{rideId}/status", middleware.Protect(http.HandlerFunc(controllers.UpdateRideStatus)))

	return mux
}
