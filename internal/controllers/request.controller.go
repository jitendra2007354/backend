package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

func RequestRide(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		PickupLocation  models.GeoPoint `json:"pickupLocation"`
		DropoffLocation models.GeoPoint `json:"dropoffLocation"`
		VehicleType     string          `json:"vehicleType"`
		Fare            float64         `json:"fare"`
		Distance        float64         `json:"distance"`
		Duration        float64         `json:"duration"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	ride, err := services.CreateRideService(user.ID, req.PickupLocation, req.DropoffLocation, req.VehicleType, req.Fare, req.Distance, req.Duration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ride)
}

func GetRideDetails(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	ride, err := services.GetRideByID(uint(id))
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(ride)
}

func CancelRideRequest(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := services.HandleRideCancellation(uint(id), "Customer"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Cancelled"})
}