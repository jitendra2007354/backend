package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func RatingRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /", middleware.Protect(http.HandlerFunc(controllers.CreateRatingController)))
	mux.Handle("GET /{userId}", middleware.Protect(http.HandlerFunc(controllers.GetRatingsController)))

	return mux
}
