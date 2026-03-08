package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func SponsorRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /login", controllers.SponsorLogin)
	mux.Handle("GET /notifications", middleware.ProtectSponsor(http.HandlerFunc(controllers.GetSponsorHistory)))
	mux.Handle("POST /notifications", middleware.ProtectSponsor(http.HandlerFunc(controllers.SendSponsorNotification)))
	mux.Handle("DELETE /notifications/{id}", middleware.ProtectSponsor(http.HandlerFunc(controllers.DeleteSponsorNotification)))
	mux.Handle("POST /banner", middleware.ProtectSponsor(http.HandlerFunc(controllers.UploadCampaignBanner)))
	mux.Handle("GET /gam-token", middleware.ProtectSponsor(http.HandlerFunc(controllers.GetGamToken)))

	return mux
}