package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func ChatRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /send", middleware.Protect(http.HandlerFunc(controllers.SendChatMessage)))
	mux.Handle("GET /messages/{rideId}", middleware.Protect(http.HandlerFunc(controllers.GetChatMessages)))

	return mux
}
