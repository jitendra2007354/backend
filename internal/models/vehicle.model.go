package models

import (
	"time"
)

// Vehicle represents the vehicle model
type Vehicle struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"not null" json:"userId"`
	User            User      `gorm:"foreignKey:UserID" json:"driver,omitempty"`
	VehicleNumber   string    `gorm:"unique;not null" json:"vehicleNumber"`
	VehicleModel    string    `gorm:"not null" json:"vehicleModel"`
	VehicleType     string    `gorm:"type:enum('Bike','Auto','Car4Seater','Car6Seater');not null" json:"vehicleType"`
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
