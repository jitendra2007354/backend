package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func AuthRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /login", controllers.Login)
	mux.HandleFunc("POST /admin-login", controllers.AdminLogin)
	mux.Handle("GET /me", middleware.Protect(http.HandlerFunc(controllers.Me)))

	return mux
}
