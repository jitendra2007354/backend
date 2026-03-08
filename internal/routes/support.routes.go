package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func SupportRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /tickets", middleware.Protect(http.HandlerFunc(controllers.SubmitSupportTicket)))
	
	return mux
}
