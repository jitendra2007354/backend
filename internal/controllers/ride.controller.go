package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

// writeRideJsonError is a helper to format error responses as JSON.
func writeRideJsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func CreateRide(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRideJsonError(w, "Invalid body", http.StatusBadRequest)
		return
	}

	pickup := parseGeoPointFromMap(req["pickupLocation"])
	dropoff := parseGeoPointFromMap(req["dropoffLocation"])

	vType, _ := req["vehicleType"].(string)

	fare := 0.0
	if f, ok := req["fare"].(float64); ok {
		fare = f
	} else if fStr, ok := req["fare"].(string); ok {
		fare, _ = strconv.ParseFloat(fStr, 64)
	}

	dist := 0.0
	if d, ok := req["distance"].(float64); ok {
		dist = d
	}

	dur := 0.0
	if d, ok := req["duration"].(float64); ok {
		dur = d
	}

	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	ride, err := services.CreateRideService(user.ID, pickup, dropoff, vType, fare, dist, dur)
	if err != nil {
		writeRideJsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ride)
}

func GetAvailableRides(w http.ResponseWriter, r *http.Request) {
	rides, err := services.GetPendingRides()
	if err != nil {
		writeRideJsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rides)
}

func GetRideByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)

	ride, err := services.GetRideWithDetails(uint(id))
	if err != nil {
		writeRideJsonError(w, "Ride not found", http.StatusNotFound)
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
		writeRideJsonError(w, "Ride not found", http.StatusNotFound)
		return
	}

	if ride.Status == "assigning" {
		err = services.HandleDriverAcceptChain(uint(id), user.ID)
	} else if ride.Status == "pending" {
		err = services.DriverAcceptInstant(uint(id), user.ID)
	} else {
		writeRideJsonError(w, "Ride not available", http.StatusBadRequest)
		return
	}

	if err != nil {
		writeRideJsonError(w, err.Error(), http.StatusInternalServerError)
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
		writeRideJsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func CancelRide(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	if err := services.HandleRideCancellation(uint(id), user.UserType); err != nil {
		writeRideJsonError(w, err.Error(), http.StatusInternalServerError)
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
		writeRideJsonError(w, err.Error(), http.StatusBadRequest)
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
		writeRideJsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(ride)
}

func GetActiveRide(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	ride, err := services.GetDriverActiveRide(user.ID)
	if err != nil {
		writeRideJsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ride == nil {
		// No active ride for this driver
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(nil)
		return
	}
	json.NewEncoder(w).Encode(ride)
}
