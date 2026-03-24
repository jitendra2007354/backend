package services

import (
	"errors"
	"time"

	"spark/internal/models"

	"gorm.io/gorm"
)

func CreateBid(rideID, userID uint, amount float64, customerID uint) (*models.Bid, error) {
	var bid models.Bid
	err := DB.Transaction(func(tx *gorm.DB) error {
		var driver models.Driver
		if err := tx.Where("user_id = ?", userID).Preload("User").First(&driver).Error; err != nil {
			return errors.New("driver not found")
		}

		// Check active ride
		var activeRide models.Ride
		if err := tx.Where("driver_id = ? AND status IN ?", driver.ID, []string{"accepted", "arrived", "in-progress"}).First(&activeRide).Error; err == nil {
			return errors.New("cannot bid while on active ride")
		}

		var ride models.Ride
		if err := tx.First(&ride, rideID).Error; err != nil || ride.Status != "pending" {
			return errors.New("ride not available")
		}

		now := time.Now()
		if ride.CurrentDriverID != nil && *ride.CurrentDriverID != driver.ID && ride.OfferExpiresAt != nil && ride.OfferExpiresAt.After(now) {
			return errors.New("ride is currently locked by another driver")
		}

		// Upsert bid
		if err := tx.Where("ride_id = ? AND driver_id = ?", rideID, driver.ID).First(&bid).Error; err == nil {
			bid.Amount = amount
			tx.Save(&bid)
		} else {
			bid = models.Bid{RideID: rideID, DriverID: driver.ID, Amount: amount}
			tx.Create(&bid)
		}

		// Lock ride to this driver for 5 minutes
		expiresAt := now.Add(5 * time.Minute)
		ride.CurrentDriverID = &driver.ID
		ride.OfferExpiresAt = &expiresAt
		tx.Save(&ride)

		// Notify customer
		SendMessageToUser(customerID, "bid-new", map[string]interface{}{
			"rideId": rideID,
			"amount": amount,
			"driver": driver,
		})
		return nil
	})
	return &bid, err
}

func GetBidsForRide(rideID uint) ([]models.Bid, error) {
	var bids []models.Bid
	err := DB.Where("ride_id = ?", rideID).Preload("Driver.User").Order("amount ASC").Find(&bids).Error
	return bids, err
}

func AcceptBid(bidID, customerID uint) (map[string]string, error) {
	err := DB.Transaction(func(tx *gorm.DB) error {
		var bid models.Bid
		if err := tx.First(&bid, bidID).Error; err != nil {
			return errors.New("bid not found")
		}

		var ride models.Ride
		if err := tx.First(&ride, bid.RideID).Error; err != nil {
			return errors.New("ride not found")
		}

		if ride.CustomerID != customerID {
			return errors.New("unauthorized")
		}
		if ride.Status != "pending" {
			return errors.New("ride not pending")
		}

		bid.IsAccepted = true
		tx.Save(&bid)

		ride.Status = "accepted"
		ride.DriverID = &bid.DriverID
		ride.Fare = bid.Amount
		ride.CurrentDriverID = nil
		tx.Save(&ride)

		ProcessDailyFee(bid.DriverID, tx)

		var driver models.Driver
		tx.First(&driver, bid.DriverID)
		SendMessageToUser(driver.UserID, "bid-accepted", map[string]interface{}{"rideId": ride.ID})

		return nil
	})
	return map[string]string{"message": "Bid accepted"}, err
}

func DriverAcceptInstant(rideID, userID uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var driver models.Driver
		if err := tx.Where("user_id = ?", userID).First(&driver).Error; err != nil {
			return errors.New("driver not found")
		}

		var activeRide models.Ride
		if err := tx.Where("driver_id = ? AND status IN ?", driver.ID, []string{"accepted", "arrived", "in-progress"}).First(&activeRide).Error; err == nil {
			return errors.New("driver busy")
		}

		var ride models.Ride
		if err := tx.First(&ride, rideID).Error; err != nil || ride.Status != "pending" {
			return errors.New("ride not available")
		}

		now := time.Now()
		if ride.CurrentDriverID != nil && *ride.CurrentDriverID != driver.ID && ride.OfferExpiresAt != nil && ride.OfferExpiresAt.After(now) {
			return errors.New("ride is currently locked by another driver")
		}

		ride.Status = "accepted"
		ride.DriverID = &driver.ID
		ride.CurrentDriverID = nil
		tx.Save(&ride)

		ProcessDailyFee(driver.ID, tx)

		SendMessageToUser(ride.CustomerID, "ride_confirmed", map[string]interface{}{
			"rideId":  ride.ID,
			"message": "Driver accepted",
		})
		SendMessageToAdminRoom("ride_updated", map[string]interface{}{
			"rideId": ride.ID,
			"status": "accepted",
		})

		return nil
	})
}

func CounterBid(bidID, customerID uint, amount float64) error {
	var bid models.Bid
	if err := DB.Preload("Ride").First(&bid, bidID).Error; err != nil {
		return errors.New("bid not found")
	}

	if bid.Ride.CustomerID != customerID {
		return errors.New("unauthorized")
	}

	SendMessageToUser(bid.DriverID, "counter-bid", map[string]interface{}{
		"bidId":  bidID,
		"amount": amount,
	})
	return nil
}

func AcceptCounterBid(bidID, driverID uint, amount float64) (*models.Ride, error) {
	var bid models.Bid
	if err := DB.First(&bid, bidID).Error; err != nil {
		return nil, errors.New("bid not found")
	}

	var ride models.Ride
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&ride, bid.RideID).Error; err != nil {
			return err
		}
		if ride.Status != "pending" {
			return errors.New("ride not pending")
		}

		// Update bid
		bid.Amount = amount
		bid.IsAccepted = true
		if err := tx.Save(&bid).Error; err != nil {
			return err
		}

		// Update ride
		ride.Status = "accepted"
		ride.DriverID = &driverID
		ride.CurrentDriverID = nil
		ride.Fare = amount
		if err := tx.Save(&ride).Error; err != nil {
			return err
		}

		ProcessDailyFee(driverID, tx)
		return nil
	})

	if err != nil {
		return nil, err
	}

	SendMessageToUser(ride.CustomerID, "bid-accepted", map[string]interface{}{"rideId": ride.ID, "fare": amount})
	return &ride, nil
}
