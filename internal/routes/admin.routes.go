package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func AdminRoutes() http.Handler {
	mux := http.NewServeMux()

	// Helper to chain Protect and IsAdmin
	adminAuth := func(h http.HandlerFunc) http.Handler {
		return middleware.Protect(middleware.IsAdmin(h))
	}

	// AI Assistant
	mux.Handle("POST /ai/generate", adminAuth(controllers.GetAIAssistantResponse))

	// Config & Users
	mux.Handle("GET /config", adminAuth(controllers.GetSystemConfig))
	mux.Handle("POST /config", adminAuth(controllers.UpdateSystemConfig))
	mux.Handle("GET /users", adminAuth(controllers.GetAllUsers))
	mux.Handle("POST /users", adminAuth(controllers.CreateUser))

	// Other Admin Routes
	mux.Handle("GET /tickets", adminAuth(controllers.GetTickets))
	mux.Handle("GET /drivers/locations", adminAuth(controllers.GetDriverLocations))
	mux.Handle("GET /notifications/history", adminAuth(controllers.GetNotificationHistory))
	mux.Handle("POST /notifications", adminAuth(controllers.SendNotification))
	mux.Handle("POST /notifications-v2", adminAuth(controllers.CreateRealtimeNotificationHandler))

	return mux
}
