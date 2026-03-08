package services

import (
	"fmt"
	"spark/internal/models"
)

func UpdateDriverLocation(driverID uint, lat, lng float64, shouldPersist bool) {
	if shouldPersist {
		// GORM upsert
		loc := models.DriverLocation{
			DriverID: driverID,
			Lat:      lat,
			Lng:      lng,
		}
		DB.Save(&loc)
	}
	// Broadcast logic
	SendMessageToAdminRoom("driver_location_updated", map[string]interface{}{"driverId": driverID, "lat": lat, "lng": lng})
}

func FindNearbyAvailableDrivers(lat, lng float64, vehicleType string, city string) ([]models.Driver, error) {
	var drivers []models.Driver
	// Using raw SQL for spatial query as GORM doesn't support ST_Distance_Sphere natively without plugins
	// This assumes MySQL 5.7+ or MariaDB
	err := DB.Raw(`
		SELECT d.* 
		FROM drivers d
		JOIN driver_locations dl ON d.id = dl.driver_id
		JOIN users u ON d.user_id = u.id
		WHERE d.vehicle_type = ? 
		AND u.is_online = true 
		AND u.is_blocked = false
		AND ST_Distance_Sphere(POINT(dl.lng, dl.lat), ST_GeomFromText(?)) <= 5000
	`, vehicleType, fmt.Sprintf("POINT(%f %f)", lng, lat)).Scan(&drivers).Error
	return drivers, err
}

func GetAllOnlineDriverLocations() ([]map[string]interface{}, error) {
	var results []struct {
		DriverID uint
		Lat      float64
		Lng      float64
	}
	// Simple select since we now have lat/lng columns
	err := DB.Table("driver_locations").Select("driver_id, lat, lng").Scan(&results).Error

	if err != nil {
		return nil, err
	}

	mapped := make([]map[string]interface{}, len(results))
	for i, r := range results {
		mapped[i] = map[string]interface{}{
			"driverId": r.DriverID,
			"lat":      r.Lat,
			"lng":      r.Lng,
		}
	}
	return mapped, nil
}

// RemoveDriverLocation deletes any stored location for a driver.
func RemoveDriverLocation(driverID uint) error {
	return DB.Where("driver_id = ?", driverID).Delete(&models.DriverLocation{}).Error
}