package services

import (
	"errors"
	"fmt"
	"spark/internal/models"
	"strings"
)

func getStringField(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key].(string); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
		if val, ok := data[key]; ok && val != nil {
			return strings.TrimSpace(fmt.Sprintf("%v", val))
		}
	}
	return ""
}

func AddVehicleForDriver(userID uint, data map[string]interface{}) (*models.Vehicle, error) {
	var driver models.Driver
	if err := DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		return nil, errors.New("driver profile not found")
	}

	vehicleNumber := getStringField(data, "vehicleNumber", "number")
	vehicleModel := getStringField(data, "vehicleModel", "model")
	vehicleType := getStringField(data, "vehicleType", "type")
	rcPhoto := getStringField(data, "rcPhotoUrl", "rcPhoto")
	licensePhoto := getStringField(data, "licensePhotoUrl", "licensePhoto")
	rcNumber := getStringField(data, "rcNumber", "rc_number")

	if vehicleNumber == "" || vehicleModel == "" || vehicleType == "" || rcNumber == "" {
		return nil, errors.New("vehicleNumber, vehicleModel, vehicleType, and rcNumber are required")
	}

	vehicle := models.Vehicle{
		UserID:          userID,
		DriverID:        driver.ID,
		VehicleNumber:   vehicleNumber,
		VehicleModel:    vehicleModel,
		VehicleType:     vehicleType,
		RCNumber:        rcNumber,
		RCPhotoURL:      rcPhoto,
		LicensePhotoURL: licensePhoto,
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

	// Sync the selected vehicle details to the Drivers table.
	// This ensures the matching algorithm (which looks at drivers table)
	// sees the correct vehicle type and number.
	DB.Model(&models.Driver{}).Where("user_id = ?", driverID).Updates(map[string]interface{}{
		"vehicle_type":   vehicle.VehicleType,
		"vehicle_model":  vehicle.VehicleModel,
		"vehicle_number": vehicle.VehicleNumber,
		"rc_number":      vehicle.RCNumber,
	})

	return &vehicle, nil
}

func DeleteVehicleService(driverID, vehicleID uint) error {
	result := DB.Model(&models.Vehicle{}).Where("id = ? AND user_id = ?", vehicleID, driverID).Update("is_deleted", true)
	if result.RowsAffected == 0 {
		return errors.New("vehicle not found")
	}
	return nil
}
