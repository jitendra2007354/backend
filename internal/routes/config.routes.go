package routes

import (
	"net/http"
	"spark/internal/controllers"
)

func ConfigRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /", controllers.UpdateConfig)
	mux.HandleFunc("GET /", controllers.GetConfig)

	return mux
}
