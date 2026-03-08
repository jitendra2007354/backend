package models

import (
	"time"
)

// Driver represents the driver model
type Driver struct {
	ID                     uint     `gorm:"primaryKey" json:"id"`
	UserID                 uint     `gorm:"unique;not null" json:"userId"`
	User                   User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	DriverLicenseNumber    string   `gorm:"unique;not null" json:"driverLicenseNumber"`
	DriverLicensePhotoURL  string   `gorm:"not null" json:"driverLicensePhotoUrl"`
	VehicleModel           string   `gorm:"not null" json:"vehicleModel"`
	VehicleNumber          string   `gorm:"unique;not null" json:"vehicleNumber"`
	VehicleType            string   `gorm:"type:enum('Bike','Auto','Car','Car 7-Seater');not null" json:"vehicleType"`
	RCPhotoURL             string   `gorm:"not null" json:"rcPhotoUrl"`
	IsApproved             bool     `gorm:"default:false" json:"isApproved"`
	AverageRating          float64  `gorm:"type:decimal(3,2);default:5.00" json:"averageRating"`
	CurrentLat             *float64 `json:"currentLat,omitempty"`
	CurrentLng             *float64 `json:"currentLng,omitempty"`
	OutstandingPlatformFee float64  `gorm:"default:0" json:"outstandingPlatformFee"`
	LastDailyFeeChargedAt  *time.Time `json:"lastDailyFeeChargedAt,omitempty"`
	PlatformAccessExpiry   *time.Time `json:"platformAccessExpiry,omitempty"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

func (Driver) TableName() string {
	return "Drivers"
}
