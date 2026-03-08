package services

import (
	"errors"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"spark/internal/models"
)

var validTransitions = map[string][]string{
	"accepted":    {"arrived"},
	"arrived":     {"in-progress"},
	"in-progress": {"completed"},
}

func CreateRideService(customerID uint, pickup, dropoff models.GeoPoint, vehicleType string, fare float64, dist, dur float64) (*models.Ride, error) {
	// Convert GeoPoint to JSON for storage
	pBytes, _ := json.Marshal(pickup)
	pickupJSON := datatypes.JSON(pBytes)
	dBytes, _ := json.Marshal(dropoff)
	dropoffJSON := datatypes.JSON(dBytes)

	ride := models.Ride{
		CustomerID:          customerID,
		PickupLocation:      pickupJSON,
		DestinationLocation: dropoffJSON,
		VehicleType:         vehicleType,
		Fare:                fare,
		Status:              "pending",
		Distance:            &dist,
		Duration:            &dur,
	}

	if err := DB.Create(&ride).Error; err != nil {
		return nil, err
	}

	// Broadcast to nearby drivers
	if len(pickup.Coordinates) == 2 {
		lat, lng := pickup.Coordinates[1], pickup.Coordinates[0]
		drivers, _ := FindNearbyAvailableDrivers(lat, lng, vehicleType, "")
		for _, d := range drivers {
			SendMessageToUser(d.UserID, "new_ride_request", map[string]interface{}{"ride": ride})
		}
	}

	return &ride, nil
}

func GetRideByID(rideID uint) (*models.Ride, error) {
	var ride models.Ride
	if err := DB.First(&ride, rideID).Error; err != nil {
		return nil, err
	}
	return &ride, nil
}

func GetRideWithDetails(rideID uint) (*models.Ride, error) {
	var ride models.Ride
	err := DB.Preload("Customer").Preload("Driver.User").First(&ride, rideID).Error
	return &ride, err
}

func ConfirmRideService(rideID, customerID uint) (*models.Ride, error) {
	ride, err := GetRideWithDetails(rideID)
	if err != nil {
		return nil, err
	}

	if ride.CustomerID != customerID {
		return nil, errors.New("unauthorized")
	}
	if ride.Status != "accepted" {
		return nil, errors.New("ride cannot be confirmed")
	}

	if ride.DriverID != nil {
		var driver models.Driver
		if err := DB.Where("id = ?", *ride.DriverID).First(&driver).Error; err == nil {
			SendMessageToUser(driver.UserID, "ride_confirmed", map[string]interface{}{
				"rideId":  ride.ID,
				"message": "Customer confirmed.",
			})
		}
	}
	return ride, nil
}

func UpdateRideStatusService(rideID, userID uint, newStatus string) (*models.Ride, error) {
	var ride models.Ride
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&ride, rideID).Error; err != nil {
			return err
		}

		var driver models.Driver
		if err := tx.Where("user_id = ?", userID).First(&driver).Error; err != nil {
			return errors.New("driver profile not found")
		}

		if ride.DriverID == nil || *ride.DriverID != driver.ID {
			return errors.New("unauthorized")
		}

		allowed, ok := validTransitions[ride.Status]
		isValid := false
		if ok {
			for _, s := range allowed {
				if s == newStatus {
					isValid = true
					break
				}
			}
		}
		if !isValid {
			return fmt.Errorf("invalid transition from %s to %s", ride.Status, newStatus)
		}

		ride.Status = newStatus
		if err := tx.Save(&ride).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	SendMessageToUser(ride.CustomerID, "ride_status_updated", map[string]interface{}{
		"rideId": ride.ID,
		"status": newStatus,
	})

	return &ride, nil
}

func GetCustomerRideHistory(customerID uint) ([]models.Ride, error) {
	var rides []models.Ride
	err := DB.Where("customer_id = ?", customerID).Order("created_at DESC").Find(&rides).Error
	return rides, err
}

func GetDriverRideHistory(userID uint) ([]models.Ride, error) {
	var driver models.Driver
	if err := DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		return []models.Ride{}, nil
	}
	var rides []models.Ride
	err := DB.Where("driver_id = ?", driver.ID).Order("created_at DESC").Find(&rides).Error
	return rides, err
}

func GetPendingRides() ([]models.Ride, error) {
	var rides []models.Ride
	err := DB.Where("status = ?", "pending").Preload("Customer").Order("created_at DESC").Find(&rides).Error
	return rides, err
}
