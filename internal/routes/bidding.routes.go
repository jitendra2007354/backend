package routes

import (
	"net/http"
	"spark/internal/controllers"
	"spark/internal/middleware"
)

func BiddingRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /", middleware.Protect(http.HandlerFunc(controllers.CreateBidController)))
	mux.Handle("GET /{rideId}", middleware.Protect(http.HandlerFunc(controllers.GetBidsController)))
	mux.Handle("POST /accept", middleware.Protect(http.HandlerFunc(controllers.AcceptBidController)))
	mux.Handle("POST /counter", middleware.Protect(http.HandlerFunc(controllers.CounterBidController)))
	mux.Handle("POST /accept-counter", middleware.Protect(http.HandlerFunc(controllers.AcceptCounterBidController)))

	return mux
}
