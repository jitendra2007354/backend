package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

func CreateRatingController(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		RideID  uint   `json:"rideId"`
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	ride, err := services.GetRideByID(req.RideID)
	if err != nil {
		http.Error(w, "Ride not found", http.StatusNotFound)
		return
	}

	ratedUserType := "customer"
	if ride.CustomerID == user.ID {
		ratedUserType = "driver"
	}

	rating, err := services.SubmitRating(req.RideID, user.ID, ratedUserType, req.Rating, req.Comment)
	json.NewEncoder(w).Encode(rating)
}

func GetRatingsController(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.PathValue("userId"))
	ratings, _ := services.GetRatingsForUser(uint(userID))
	json.NewEncoder(w).Encode(ratings)
}
