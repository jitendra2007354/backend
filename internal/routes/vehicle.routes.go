package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func VehicleRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /", middleware.Protect(http.HandlerFunc(controllers.AddVehicle)))
	mux.Handle("GET /", middleware.Protect(http.HandlerFunc(controllers.GetVehicles)))
	mux.Handle("PATCH /{vehicleId}/default", middleware.Protect(http.HandlerFunc(controllers.SetDefaultVehicle)))
	mux.Handle("DELETE /{vehicleId}", middleware.Protect(http.HandlerFunc(controllers.DeleteVehicle)))

	return mux
}
