package routes

import (
	"net/http"
	"spark/internal/controllers"
)

func UserRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", controllers.RegisterUser)

	return mux
}
