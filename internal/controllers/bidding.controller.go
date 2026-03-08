package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

func CreateBidController(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	if user.UserType != "Driver" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		RideID uint    `json:"rideId"`
		Amount float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Get Ride to find customer ID
	ride, err := services.GetRideByID(req.RideID)
	if err != nil || ride.Status != "pending" {
		http.Error(w, "Ride not available", http.StatusBadRequest)
		return
	}

	bid, err := services.CreateBid(req.RideID, user.ID, req.Amount, ride.CustomerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bid)
}

func GetBidsController(w http.ResponseWriter, r *http.Request) {
	rideID, _ := strconv.Atoi(r.PathValue("rideId"))
	bids, err := services.GetBidsForRide(uint(rideID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bids)
}

func AcceptBidController(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	if user.UserType != "Customer" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		BidID uint `json:"bidId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := services.AcceptBid(req.BidID, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func CounterBidController(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		BidID  uint    `json:"bidId"`
		Amount float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := services.CounterBid(req.BidID, user.ID, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Counter-offer sent to driver."})
}

func AcceptCounterBidController(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		BidID  uint    `json:"bidId"`
		Amount float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := services.AcceptCounterBid(req.BidID, user.ID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func DriverAcceptInstantController(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		RideID uint `json:"rideId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := services.DriverAcceptInstant(req.RideID, user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Ride accepted"})
}
