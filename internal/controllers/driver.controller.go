package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"spark/internal/database"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
	"strings"
	"time"
)

func GetDriverRides(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	rides, err := services.GetDriverRideHistory(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rides)
}

func GetNearbyDrivers(w http.ResponseWriter, r *http.Request) {
	lat, latErr := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, lngErr := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if latErr != nil || lngErr != nil {
		http.Error(w, "Invalid lat/lng parameters", http.StatusBadRequest)
		return
	}

	vType := r.URL.Query().Get("vehicleType")
	if vType == "" {
		vType = r.URL.Query().Get("type")
	}

	drivers, err := services.FindNearbyDrivers(lat, lng, vType)
	if err != nil {
		log.Printf("[Driver] GetNearbyDrivers error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Failed to fetch nearby drivers",
			"error":   err.Error(),
		})
		return
	}
	if drivers == nil {
		drivers = []models.Driver{}
	}
	json.NewEncoder(w).Encode(drivers)
}

func BlockDriver(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := services.BlockDriver(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Blocked"})
}

func UnblockDriver(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := services.UnblockDriver(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Unblocked"})
}

func PollRideRequests(w http.ResponseWriter, r *http.Request) {
	// Ignore old "ghost" requests that were created more than 30 minutes ago
	expirationTime := time.Now().Add(-30 * time.Minute)

	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok {
		json.NewEncoder(w).Encode(nil)
		return
	}

	var driver models.Driver
	if err := database.DB.Where("user_id = ?", user.ID).First(&driver).Error; err != nil {
		json.NewEncoder(w).Encode(nil)
		return
	}

	var activeTypes []string
	if user.ActiveVehicleTypes != nil && string(user.ActiveVehicleTypes) != "" && string(user.ActiveVehicleTypes) != "[]" {
		json.Unmarshal(user.ActiveVehicleTypes, &activeTypes)
	} else {
		activeTypes = append(activeTypes, driver.VehicleType)
	}

	now := time.Now()
	query := database.DB.Where("status = ? AND (distance IS NULL OR distance >= 0) AND created_at > ? AND (current_driver_id IS NULL OR current_driver_id = ? OR offer_expires_at < ?)", "pending", expirationTime, driver.ID, now).Order("created_at DESC").Preload("Customer")

	ignored := r.URL.Query().Get("ignored")
	if ignored != "" {
		ids := strings.Split(ignored, ",")
		if len(ids) > 0 {
			query = query.Where("id NOT IN ?", ids)
		}
	}

	var rides []models.Ride
	if err := query.Limit(20).Find(&rides).Error; err != nil {
		json.NewEncoder(w).Encode(nil)
		return
	}

	var matchedRide *models.Ride
	for i := range rides {
		rType := rides[i].VehicleType
		if rType == "4 Seater" || rType == "Car" {
			rType = "Car 4-Seater"
		} else if rType == "6 Seater" {
			rType = "Car 7-Seater"
		}

		matched := false
		for _, at := range activeTypes {
			if at == rType || at == rides[i].VehicleType {
				matched = true
				break
			}
		}

		if matched {
			matchedRide = &rides[i]
			break
		}
	}

	if matchedRide == nil {
		json.NewEncoder(w).Encode(nil)
		return
	}

	ride := *matchedRide

	// perform transformation matching TS version
	resp := map[string]interface{}{
		"id": ride.ID,
		"customer": map[string]interface{}{
			"id":     ride.Customer.ID,
			"name":   fmt.Sprintf("%s %s", ride.Customer.FirstName, ride.Customer.LastName),
			"rating": ride.Customer.AverageRating,
			"pfp":    ride.Customer.PFP,
		},
		"pickupLocation":  ride.PickupLocation,
		"dropoffLocation": ride.DestinationLocation,
		"fare":            ride.Fare,
		"distance":        ride.Distance,
		"duration":        ride.Duration,
		"vehicleType":     ride.VehicleType,
		"expiresAt": func() int64 {
			if ride.OfferExpiresAt != nil {
				return ride.OfferExpiresAt.UnixNano() / int64(time.Millisecond)
			}
			return time.Now().Add(30*time.Second).UnixNano() / int64(time.Millisecond)
		}(),
	}
	json.NewEncoder(w).Encode(resp)
}

func ToggleDriverStatus(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	newStatus := !user.IsOnline
	services.SetDriverOnlineStatus(user.ID, newStatus)
	json.NewEncoder(w).Encode(map[string]bool{"isOnline": newStatus})
}

func SetDriverStatus(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var payload struct {
		IsOnline bool `json:"isOnline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	services.SetDriverOnlineStatus(user.ID, payload.IsOnline)
	json.NewEncoder(w).Encode(map[string]bool{"isOnline": payload.IsOnline})
}
