package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

func CreateRide(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PickupLocation  models.GeoPoint `json:"pickupLocation"`
		DropoffLocation models.GeoPoint `json:"dropoffLocation"`
		VehicleType     string          `json:"vehicleType"`
		Fare            float64         `json:"fare,string"` // Handle string or number input
		Distance        float64         `json:"distance"`
		Duration        float64         `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	ride, err := services.CreateRideService(user.ID, req.PickupLocation, req.DropoffLocation, req.VehicleType, req.Fare, req.Distance, req.Duration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ride)
}

func GetAvailableRides(w http.ResponseWriter, r *http.Request) {
	rides, err := services.GetPendingRides()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rides)
}

func GetRideByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)

	ride, err := services.GetRideWithDetails(uint(id))
	if err != nil {
		http.Error(w, "Ride not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(ride)
}

func AcceptRide(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	ride, err := services.GetRideByID(uint(id))
	if err != nil {
		http.Error(w, "Ride not found", http.StatusNotFound)
		return
	}

	if ride.Status == "assigning" {
		err = services.HandleDriverAcceptChain(uint(id), user.ID)
	} else if ride.Status == "pending" {
		err = services.DriverAcceptInstant(uint(id), user.ID)
	} else {
		http.Error(w, "Ride not available", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedRide, _ := services.GetRideByID(uint(id))
	json.NewEncoder(w).Encode(updatedRide)
}

func RejectRide(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	if err := services.HandleDriverReject(uint(id), user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func CancelRide(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	if err := services.HandleRideCancellation(uint(id), user.UserType); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func ConfirmRide(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("rideId")
	id, _ := strconv.Atoi(idStr)
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	ride, err := services.ConfirmRideService(uint(id), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(ride)
}

func UpdateRideStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("rideId")
	id, _ := strconv.Atoi(idStr)
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	ride, err := services.UpdateRideStatusService(uint(id), user.ID, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(ride)
}
