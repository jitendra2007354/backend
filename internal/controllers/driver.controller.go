package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"spark/internal/database"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
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
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	vType := r.URL.Query().Get("vehicleType")

	drivers, err := services.FindNearbyDrivers(lat, lng, vType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	var ride models.Ride
	if err := database.DB.Where("status = ?", "pending").Order("created_at DESC").Preload("Customer").First(&ride).Error; err != nil {
		json.NewEncoder(w).Encode(nil)
		return
	}

	// perform transformation matching TS version
	resp := map[string]interface{}{
		"id": ride.ID,
		"customer": map[string]interface{}{
			"id":    ride.Customer.ID,
			"name":  fmt.Sprintf("%s %s", ride.Customer.FirstName, ride.Customer.LastName),
			"rating": ride.Customer.AverageRating,
			"pfp":   ride.Customer.PFP,
		},
		"pickupLocation": ride.PickupLocation,
		"dropoffLocation": ride.DestinationLocation,
		"fare": ride.Fare,
		"distance": ride.Distance,
		"duration": ride.Duration,
		"vehicleType": ride.VehicleType,
		"expiresAt": func() int64 {
			if ride.OfferExpiresAt != nil {
				return ride.OfferExpiresAt.UnixNano() / int64(time.Millisecond)
			}
			return time.Now().Add(30 * time.Second).UnixNano() / int64(time.Millisecond)
		}(),
	}
	json.NewEncoder(w).Encode(resp)
}
