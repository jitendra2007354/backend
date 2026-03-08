package services

import (
	"math/rand"
	"spark/internal/models"
)

func GetDistanceAndDuration(origin, dest models.GeoPoint) (float64, float64, error) {
	// Mocked values
	distance := float64(5000 + rand.Intn(20000))
	duration := distance/15 + float64(rand.Intn(600))
	return distance, duration, nil
}