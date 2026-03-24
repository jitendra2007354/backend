package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"spark/internal/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var validTransitions = map[string][]string{
	"accepted":    {"arrived", "in-progress", "completed"},
	"arrived":     {"in-progress", "completed"},
	"in-progress": {"completed"},
}

func validateCustomerMinFare(vehicleType string, dist float64, fare float64) error {
	distanceKm := dist / 1000.0
	if distanceKm <= 0 {
		distanceKm = 1.0
	}

	baseMinFare := 50.0
	minBidPerKm := 8.0

	if config, err := GetApplicableConfig(); err == nil {
		switch vehicleType {
		case "Bike":
			if config.BaseFareBike > 0 {
				baseMinFare = config.BaseFareBike
			} else {
				baseMinFare = config.BaseFare * 0.5
			}
			if config.MinBidPerKmBike > 0 {
				minBidPerKm = config.MinBidPerKmBike
			} else {
				minBidPerKm = config.MinBidPerKm * 0.6
			}
		case "Auto":
			if config.BaseFareAuto > 0 {
				baseMinFare = config.BaseFareAuto
			} else {
				baseMinFare = config.BaseFare * 0.7
			}
			if config.MinBidPerKmAuto > 0 {
				minBidPerKm = config.MinBidPerKmAuto
			} else {
				minBidPerKm = config.MinBidPerKm * 0.8
			}
		case "Car 7-Seater", "6 Seater":
			if config.BaseFareSUV > 0 {
				baseMinFare = config.BaseFareSUV
			} else {
				baseMinFare = config.BaseFare * 1.3
			}
			if config.MinBidPerKmSUV > 0 {
				minBidPerKm = config.MinBidPerKmSUV
			} else {
				minBidPerKm = config.MinBidPerKm * 1.2
			}
		case "Luxury":
			if config.BaseFareLuxury > 0 {
				baseMinFare = config.BaseFareLuxury
			} else {
				baseMinFare = config.BaseFare * 2.0
			}
			if config.MinBidPerKmLuxury > 0 {
				minBidPerKm = config.MinBidPerKmLuxury
			} else {
				minBidPerKm = config.MinBidPerKm * 1.8
			}
		default:
			if config.BaseFareCar > 0 {
				baseMinFare = config.BaseFareCar
			} else if config.BaseFare > 0 {
				baseMinFare = config.BaseFare
			}
			if config.MinBidPerKmCar > 0 {
				minBidPerKm = config.MinBidPerKmCar
			} else if config.MinBidPerKm > 0 {
				minBidPerKm = config.MinBidPerKm
			}
		}
	}
	minAllowed := baseMinFare
	if distMin := minBidPerKm * distanceKm; distMin > minAllowed {
		minAllowed = distMin
	}
	if fare < minAllowed {
		return fmt.Errorf("fare is below the minimum allowed limit of ₹%.0f", minAllowed)
	}
	return nil
}

func CreateRideService(customerID uint, pickup, dropoff models.GeoPoint, vehicleType string, fare float64, dist, dur float64) (*models.Ride, error) {
	if err := validateCustomerMinFare(vehicleType, dist, fare); err != nil {
		return nil, err
	}

	// Convert GeoPoint to JSON for storage
	pBytes, _ := json.Marshal(pickup)
	pickupJSON := datatypes.JSON(pBytes)
	dBytes, _ := json.Marshal(dropoff)
	dropoffJSON := datatypes.JSON(dBytes)

	draftDist := float64(-1)
	ride := models.Ride{
		CustomerID:          customerID,
		PickupLocation:      pickupJSON,
		DestinationLocation: dropoffJSON,
		VehicleType:         vehicleType,
		Fare:                fare,
		Status:              "pending",
		Distance:            &draftDist, // Use -1 to indicate draft state
		Duration:            &dur,
	}

	if err := DB.Create(&ride).Error; err != nil {
		return nil, err
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

func UpdateRideFare(rideID, customerID uint, newFare float64) (*models.Ride, error) {
	var ride models.Ride
	if err := DB.First(&ride, rideID).Error; err != nil {
		return nil, err
	}
	if ride.CustomerID != customerID {
		return nil, errors.New("unauthorized")
	}

	dist := 0.0
	if ride.Distance != nil {
		dist = *ride.Distance
	}
	if err := validateCustomerMinFare(ride.VehicleType, dist, newFare); err != nil {
		return nil, err
	}

	wasDraft := ride.Distance != nil && *ride.Distance < 0

	ride.Fare = newFare
	if wasDraft {
		validDist := float64(0)
		ride.Distance = &validDist
	}

	if err := DB.Save(&ride).Error; err != nil {
		return nil, err
	}

	if wasDraft {
		var pickup models.GeoPoint
		json.Unmarshal(ride.PickupLocation, &pickup)
		if len(pickup.Coordinates) == 2 {
			lat, lng := pickup.Coordinates[1], pickup.Coordinates[0]
			drivers, _ := FindNearbyAvailableDrivers(lat, lng, ride.VehicleType, "")
			for _, d := range drivers {
				SendMessageToUser(d.UserID, "new_ride_request", map[string]interface{}{"ride": ride})
			}
		}
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

	// If the ride is still pending, the customer is confirming a driver's bid.
	if ride.Status == "pending" {
		var bid struct {
			DriverID uint
			Amount   float64
		}
		if err := DB.Table("bids").Where("ride_id = ?", rideID).Order("amount ASC").Limit(1).Scan(&bid).Error; err == nil && bid.DriverID != 0 {
			ride.DriverID = &bid.DriverID
			ride.Fare = bid.Amount
			ride.Status = "accepted"
			DB.Save(ride)
			ProcessDailyFee(bid.DriverID, DB)
		}
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

	// Re-fetch to return fully populated data if we updated it
	if updatedRide, err := GetRideWithDetails(rideID); err == nil {
		return updatedRide, nil
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

		if newStatus == "arrived" || newStatus == "in-progress" {
			ProcessDailyFee(driver.ID, tx)
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

func GetDriverActiveRide(userID uint) (*models.Ride, error) {
	var driver models.Driver
	if err := DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		return nil, nil
	}
	var ride models.Ride
	err := DB.Where("driver_id = ? AND status IN ?", driver.ID, []string{"accepted", "arrived", "in-progress"}).First(&ride).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ride, nil
}

func GetPendingRides() ([]models.Ride, error) {
	var rides []models.Ride
	err := DB.Where("status = ?", "pending").Preload("Customer").Order("created_at DESC").Find(&rides).Error
	return rides, err
}
