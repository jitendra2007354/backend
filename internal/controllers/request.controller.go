package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

func parseGeoPointFromMap(raw interface{}) models.GeoPoint {
	if raw == nil {
		return models.GeoPoint{Type: "Point", Coordinates: []float64{0, 0}}
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return models.GeoPoint{Type: "Point", Coordinates: []float64{0, 0}}
	}

	if t, ok := m["type"].(string); ok && t == "Point" {
		if coords, ok := m["coordinates"].([]interface{}); ok && len(coords) >= 2 {
			c0, _ := coords[0].(float64)
			c1, _ := coords[1].(float64)
			return models.GeoPoint{Type: "Point", Coordinates: []float64{c0, c1}}
		}
	}

	getFloat := func(keys ...string) float64 {
		for _, k := range keys {
			if v, ok := m[k].(float64); ok {
				return v
			}
		}
		return 0
	}

	lat := getFloat("lat", "latitude", "Lat", "Latitude")
	lng := getFloat("lng", "longitude", "Lng", "Longitude")
	return models.GeoPoint{Type: "Point", Coordinates: []float64{lng, lat}}
}

func RequestRide(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
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

	ride, err := services.CreateRideService(user.ID, pickup, dropoff, vType, fare, dist, dur)
	if err != nil {
		// Use 400 Bad Request to display validation/fare errors cleanly in the frontend app
		http.Error(w, err.Error(), http.StatusBadRequest)
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
