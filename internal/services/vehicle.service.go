package services

import (
	"errors"
	"spark/internal/models"
)

func AddVehicleForDriver(driverID uint, data map[string]interface{}) (*models.Vehicle, error) {
	var user models.User
	if err := DB.First(&user, driverID).Error; err != nil || user.UserType != "Driver" {
		return nil, errors.New("driver not found")
	}

	vehicle := models.Vehicle{
		UserID:          driverID,
		VehicleNumber:   data["vehicleNumber"].(string),
		VehicleModel:    data["vehicleModel"].(string),
		VehicleType:     data["vehicleType"].(string),
		RCPhotoURL:      data["rcPhotoUrl"].(string),
		LicensePhotoURL: data["licensePhotoUrl"].(string),
		IsDefault:       false,
	}

	if err := DB.Create(&vehicle).Error; err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func ListDriverVehicles(driverID uint) ([]models.Vehicle, error) {
	var vehicles []models.Vehicle
	err := DB.Where("user_id = ? AND is_deleted = ?", driverID, false).Find(&vehicles).Error
	return vehicles, err
}

func SetDefaultVehicleService(driverID, vehicleID uint) (*models.Vehicle, error) {
	DB.Model(&models.Vehicle{}).Where("user_id = ?", driverID).Update("is_default", false)
	
	var vehicle models.Vehicle
	if err := DB.Model(&vehicle).Where("id = ? AND user_id = ? AND is_deleted = ?", vehicleID, driverID, false).Update("is_default", true).Error; err != nil {
		return nil, errors.New("vehicle not found")
	}
	DB.First(&vehicle, vehicleID)
	return &vehicle, nil
}

func DeleteVehicleService(driverID, vehicleID uint) error {
	result := DB.Model(&models.Vehicle{}).Where("id = ? AND user_id = ?", vehicleID, driverID).Update("is_deleted", true)
	if result.RowsAffected == 0 {
		return errors.New("vehicle not found")
	}
	return nil
}
