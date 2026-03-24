package services

import (
	"errors"
	"sync"
	"time"

	"spark/internal/models"

	"gorm.io/gorm"
)

var offerTimers = sync.Map{}

func HandleDriverAcceptChain(rideID, userID uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var driver models.Driver
		if err := tx.Where("user_id = ?", userID).First(&driver).Error; err != nil {
			return errors.New("driver not found")
		}

		var ride models.Ride
		if err := tx.First(&ride, rideID).Error; err != nil {
			return errors.New("ride not found")
		}

		if ride.Status != "assigning" || ride.CurrentDriverID == nil || *ride.CurrentDriverID != driver.ID {
			return errors.New("offer invalid")
		}

		if timer, ok := offerTimers.Load(rideID); ok {
			timer.(*time.Timer).Stop()
			offerTimers.Delete(rideID)
		}

		ride.Status = "accepted"
		ride.DriverID = &driver.ID
		ride.CurrentDriverID = nil
		tx.Save(&ride)

		ProcessDailyFee(driver.ID, tx)

		SendMessageToUser(ride.CustomerID, "ride_accepted", map[string]interface{}{"rideId": rideID})
		return nil
	})
}

func HandleDriverReject(rideID, userID uint) error {
	var driver models.Driver
	if err := DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		return nil
	}

	var ride models.Ride
	if err := DB.First(&ride, rideID).Error; err != nil {
		return nil
	}

	if ride.Status != "assigning" || ride.CurrentDriverID == nil || *ride.CurrentDriverID != driver.ID {
		return nil
	}

	if timer, ok := offerTimers.Load(rideID); ok {
		timer.(*time.Timer).Stop()
		offerTimers.Delete(rideID)
	}

	// Logic to find next driver would go here (simplified)
	ride.Status = "pending"
	ride.CurrentDriverID = nil
	DB.Save(&ride)

	// Trigger next assignment
	// FindAndOfferToNextDriver(rideID)
	return nil
}

func HandleRideCancellation(rideID uint, cancelledBy string) error {
	var ride models.Ride
	if err := DB.First(&ride, rideID).Error; err != nil {
		return err
	}

	if timer, ok := offerTimers.Load(rideID); ok {
		timer.(*time.Timer).Stop()
		offerTimers.Delete(rideID)
	}

	if cancelledBy == "Driver" {
		ride.Status = "cancelled"
		ride.DriverID = nil
		DB.Save(&ride)
		SendMessageToUser(ride.CustomerID, "ride_driver_cancelled", map[string]string{"message": "Driver cancelled"})
	} else {
		ride.Status = "cancelled"
		DB.Save(&ride)
		SendMessageToUser(ride.CustomerID, "ride_cancelled", nil)
		if ride.DriverID != nil {
			// Notify driver
		}
	}
	return nil
}

func HandleRideCompletion(rideID uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var ride models.Ride
		if err := tx.First(&ride, rideID).Error; err != nil {
			return err
		}

		if ride.Status != "in-progress" {
			return errors.New("ride not in progress")
		}

		ride.Status = "completed"
		if err := tx.Save(&ride).Error; err != nil {
			return err
		}

		SendMessageToUser(ride.CustomerID, "ride_completed", map[string]interface{}{"rideId": rideID})
		return nil
	})
}
