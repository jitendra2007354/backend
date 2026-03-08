package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

func AddVehicle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	vehicle, err := services.AddVehicleForDriver(user.ID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vehicle)
}

func GetVehicles(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	vehicles, _ := services.ListDriverVehicles(user.ID)
	json.NewEncoder(w).Encode(vehicles)
}

func SetDefaultVehicle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	id, _ := strconv.Atoi(r.PathValue("vehicleId"))

	vehicle, err := services.SetDefaultVehicleService(user.ID, uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(vehicle)
}

func DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	id, _ := strconv.Atoi(r.PathValue("vehicleId"))

	services.DeleteVehicleService(user.ID, uint(id))
	json.NewEncoder(w).Encode(map[string]string{"message": "Deleted"})
}
