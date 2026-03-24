package models

import (
	"time"
)

// Vehicle represents the vehicle model
type Vehicle struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	DriverID        uint      `gorm:"not null" json:"driverId"`
	Driver          Driver    `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	UserID          uint      `gorm:"not null" json:"userId"` // Keep if you need a direct link to the user as well
	VehicleNumber   string    `gorm:"unique;not null" json:"vehicleNumber"`
	VehicleModel    string    `gorm:"not null" json:"vehicleModel"`
	VehicleType     string    `gorm:"not null" json:"vehicleType"`
	RCNumber        string    `gorm:"not null" json:"rcNumber"`
	RCPhotoURL      string    `gorm:"not null" json:"rcPhotoUrl"`
	LicensePhotoURL string    `gorm:"not null" json:"licensePhotoUrl"`
	IsDefault       bool      `gorm:"default:false" json:"isDefault"`
	IsDeleted       bool      `gorm:"default:false" json:"isDeleted"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (Vehicle) TableName() string {
	return "vehicles"
}
