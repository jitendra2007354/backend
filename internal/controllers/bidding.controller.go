package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

func CreateBidController(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		RideID uint    `json:"rideId"`
		Amount float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.RideID == 0 {
		idStr := r.PathValue("rideId")
		if idStr != "" {
			if id, err := strconv.Atoi(idStr); err == nil {
				req.RideID = uint(id)
			}
		}
	}

	if req.RideID == 0 {
		http.Error(w, "rideId is required", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, "amount must be greater than zero", http.StatusBadRequest)
		return
	}

	// Customer is allowed to propose a new fare on pending ride requests.
	if user.UserType == "Customer" {
		ride, err := services.UpdateRideFare(req.RideID, user.ID, req.Amount)
		if err != nil {
			if err.Error() == "unauthorized" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(ride)
		return
	}

	if user.UserType != "Driver" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Driver creates or updates a bid response.
	ride, err := services.GetRideByID(req.RideID)
	if err != nil || ride.Status != "pending" {
		http.Error(w, "Ride not available", http.StatusBadRequest)
		return
	}

	// Validate against Maximum Allowed Bid (Same for all vehicle types)
	distanceKm := 1.0 // Default to 1.0 km to prevent calculation from collapsing to 0
	if ride.Distance != nil && *ride.Distance > 0 {
		distanceKm = *ride.Distance / 1000.0
	}

	config, _ := services.GetApplicableConfig()
	maxBidRate := config.MaxBidPerKm
	if maxBidRate <= 0 {
		maxBidRate = 30.0
	}

	baseFare := config.BaseFare
	if baseFare <= 0 {
		baseFare = 50.0
	}

	maxAllowedBid := distanceKm * maxBidRate
	// Safe floor: Ensure the max limit is never lower than double the base fare
	if maxAllowedBid < (baseFare * 2) {
		maxAllowedBid = baseFare * 2
	}

	if req.Amount > maxAllowedBid {
		http.Error(w, fmt.Sprintf("Bid amount exceeds maximum allowed limit of ₹%.2f", maxAllowedBid), http.StatusBadRequest)
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
