package services

import (
	"errors"
	"spark/internal/models"

	"gorm.io/gorm"
)

func SubmitRating(rideID, raterID uint, ratedUserType string, rating int, comment string) (*models.Rating, error) {
	var newRating models.Rating
	err := DB.Transaction(func(tx *gorm.DB) error {
		var ride models.Ride
		if err := tx.First(&ride, rideID).Error; err != nil {
			return errors.New("ride not found")
		}

		if ride.CustomerID != raterID && (ride.DriverID == nil || *ride.DriverID != raterID) {
			return errors.New("unauthorized")
		}

		if ride.Status != "completed" {
			return errors.New("ride not completed")
		}

		ratedID := ride.CustomerID
		if ratedUserType == "driver" {
			if ride.DriverID == nil {
				return errors.New("no driver")
			}
			ratedID = *ride.DriverID
		}

		newRating = models.Rating{
			RideID:  rideID,
			RaterID: raterID,
			RatedID: ratedID,
			Rating:  rating,
			Comment: &comment,
		}
		if err := tx.Create(&newRating).Error; err != nil {
			return err
		}
		return nil
	})
	return &newRating, err
}

func GetRatingsForUser(userID uint) ([]models.Rating, error) {
	var ratings []models.Rating
	err := DB.Where("rated_id = ?", userID).Preload("Rater").Find(&ratings).Error
	return ratings, err
}
