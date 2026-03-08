package services

import "math"

func CalculateFare(distance, duration float64, city, vehicleType string) (float64, map[string]interface{}, error) {
	config, _ := GetApplicableConfig()

	baseFare := config.BaseFare
	distFare := (distance / 1000) * config.PerKmRate
	timeFare := (duration / 60) * config.PerMinuteRate
	total := baseFare + distFare + timeFare

	total *= config.SurgeMultiplier
	tax := total * (config.TaxRate / 100)
	finalFare := math.Round((total+tax)*100) / 100

	return finalFare, map[string]interface{}{
		"baseFare":     baseFare,
		"distanceFare": distFare,
		"timeFare":     timeFare,
		"tax":          tax,
	}, nil
}