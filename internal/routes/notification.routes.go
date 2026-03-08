package routes

import (
	"net/http"
	"spark/internal/controllers"
)

func NotificationRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /user", controllers.SendToUser)
	mux.HandleFunc("POST /group", controllers.SendToGroup)
	mux.HandleFunc("POST /all", controllers.SendToAllUsers)

	return mux
}
