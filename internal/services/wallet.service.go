package services

import (
	"errors"
	"spark/internal/models"
)

func GetDriverWalletBalance(userID uint) (float64, error) {
	var driver models.Driver
	if err := DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		return 0, errors.New("driver not found")
	}
	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		return 0, err
	}
	return user.WalletBalance, nil
}

func TopUpWallet(userID uint, amount float64) (*models.User, error) {
	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	user.WalletBalance += amount
	if err := DB.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUserWallet(driverID uint, amount float64) (*models.User, error) {
	var driver models.Driver
	if err := DB.First(&driver, driverID).Error; err != nil {
		return nil, errors.New("driver not found")
	}
	return TopUpWallet(driver.UserID, amount)
}