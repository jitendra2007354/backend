package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func DocumentRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /upload", middleware.Protect(http.HandlerFunc(controllers.UploadDocumentController)))
	mux.Handle("POST /verify", middleware.Protect(middleware.IsAdmin(http.HandlerFunc(controllers.VerifyDocumentController))))

	return mux
}
