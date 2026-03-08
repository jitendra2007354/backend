package models

import (
	"time"
)

// GeoPoint represents a GeoJSON point
type GeoPoint struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// DriverLocation represents the driver location model
type DriverLocation struct {
	DriverID  uint    `gorm:"primaryKey"`
	Lat       float64 `gorm:"not null"`
	Lng       float64 `gorm:"not null"`
	UpdatedAt time.Time
}

func (DriverLocation) TableName() string {
	return "driver_locations"
}
