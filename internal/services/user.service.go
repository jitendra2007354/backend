package services

import "spark/internal/models"

func ApplyPenaltyToUser(userID uint, amount float64) {
	var driver models.Driver
	if err := DB.Where("user_id = ?", userID).First(&driver).Error; err == nil {
		driver.OutstandingPlatformFee += amount
		DB.Save(&driver)
		return
	}
	// Fallback to user logic if needed
}
