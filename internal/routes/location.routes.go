package routes

import (
	"net/http"
	"spark/internal/controllers"
)

func LocationRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /drivers", controllers.GetDriversLocationController)

	return mux
}
