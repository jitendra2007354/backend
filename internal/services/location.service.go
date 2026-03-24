package services

import (
	"log"
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

	// Normalize vehicle types to match between customer and driver apps
	if vehicleType == "4 Seater" || vehicleType == "Car" {
		vehicleType = "Car 4-Seater"
	} else if vehicleType == "6 Seater" {
		vehicleType = "Car 7-Seater"
	}

	var err error
	if vehicleType == "" {
		err = DB.Raw(`
			SELECT d.* 
			FROM drivers d
			JOIN driver_locations dl ON d.id = dl.driver_id
			JOIN users u ON d.user_id = u.id
			WHERE u.is_online = true 
			AND u.is_blocked = false
			AND (6371 * acos(cos(radians(?)) * cos(radians(dl.lat)) * cos(radians(dl.lng) - radians(?)) + sin(radians(?)) * sin(radians(dl.lat)))) <= 5
		`, lat, lng, lat).Scan(&drivers).Error

		if err != nil {
			log.Printf("[Location] haversine query failed, falling back to simple query: %v", err)
			err = DB.Raw(`
				SELECT d.*
				FROM drivers d
				JOIN driver_locations dl ON d.id = dl.driver_id
				JOIN users u ON d.user_id = u.id
				WHERE u.is_online = true
				AND u.is_blocked = false
			`).Scan(&drivers).Error
		}
	} else {
		searchType := "%\"" + vehicleType + "\"%"

		// Using Haversine formula for distance in km (6371 is Earth radius in km). <= 5 means within 5km.
		// This mathematical approach is highly compatible and avoids missing spatial FUNCTION errors.
		err = DB.Raw(`
			SELECT d.* 
			FROM drivers d
			JOIN driver_locations dl ON d.id = dl.driver_id
			JOIN users u ON d.user_id = u.id
			WHERE (u.active_vehicle_types LIKE ? OR ((u.active_vehicle_types IS NULL OR u.active_vehicle_types = '' OR u.active_vehicle_types = '[]') AND d.vehicle_type = ?))
			AND u.is_online = true 
			AND u.is_blocked = false
			AND (6371 * acos(cos(radians(?)) * cos(radians(dl.lat)) * cos(radians(dl.lng) - radians(?)) + sin(radians(?)) * sin(radians(dl.lat)))) <= 5
		`, searchType, vehicleType, lat, lng, lat).Scan(&drivers).Error

		if err != nil {
			log.Printf("[Location] haversine query failed, falling back to simple query: %v", err)
			err = DB.Raw(`
				SELECT d.*
				FROM drivers d
				JOIN driver_locations dl ON d.id = dl.driver_id
				JOIN users u ON d.user_id = u.id
				WHERE (u.active_vehicle_types LIKE ? OR ((u.active_vehicle_types IS NULL OR u.active_vehicle_types = '' OR u.active_vehicle_types = '[]') AND d.vehicle_type = ?))
				AND u.is_online = true
				AND u.is_blocked = false
			`, searchType, vehicleType).Scan(&drivers).Error
		}
	}

	if err != nil {
		log.Printf("[Location] nearby driver lookup failed after both queries: %v", err)
		return nil, err
	}
	return drivers, nil
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
