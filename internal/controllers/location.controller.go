package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/services"
)

func GetDriversLocationController(w http.ResponseWriter, r *http.Request) {
	locations, err := services.GetAllOnlineDriverLocations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(locations)
}
