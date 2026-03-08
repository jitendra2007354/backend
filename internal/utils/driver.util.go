package utils

import (
	"math"
)

// CalculateDistance calculates the distance between two geographical points using the Haversine formula.
// Parameters:
// lat1, lon1: Latitude and Longitude of the first point
// lat2, lon2: Latitude and Longitude of the second point
// Returns: The distance in kilometers.
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Radius of the Earth in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c // Distance in km
	return distance
}
