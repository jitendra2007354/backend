package services

import "spark/internal/models"

func FindNearbyDrivers(lat, lng float64, vehicleType string) ([]models.Driver, error) {
	return FindNearbyAvailableDrivers(lat, lng, vehicleType, "")
}

func BlockDriver(userID uint) error {
	return DB.Model(&models.User{}).Where("id = ?", userID).Update("is_blocked", true).Error
}

func UnblockDriver(userID uint) error {
	return DB.Model(&models.User{}).Where("id = ?", userID).Update("is_blocked", false).Error
}

func SetDriverOnlineStatus(userID uint, isOnline bool) {
	DB.Model(&models.User{}).Where("id = ?", userID).Update("is_online", isOnline)
}
