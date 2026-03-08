package services

import "spark/internal/models"

func FindNearbyDriversForCustomer(lat, lng float64) ([]models.Driver, error) {
	return FindNearbyAvailableDrivers(lat, lng, "", "")
}
