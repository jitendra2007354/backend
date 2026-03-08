package models

import (
	"time"
)

// Bill represents the bill model
type Bill struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	RideID         uint      `gorm:"unique;not null" json:"rideId"`
	Ride           Ride      `gorm:"foreignKey:RideID" json:"ride,omitempty"`
	BaseFare       float64   `gorm:"type:decimal(10,2);not null" json:"baseFare"`
	DistanceFare   float64   `gorm:"type:decimal(10,2);not null" json:"distanceFare"`
	TimeFare       float64   `gorm:"type:decimal(10,2);not null" json:"timeFare"`
	PlatformFee    float64   `gorm:"type:decimal(10,2);not null" json:"platformFee"`
	Taxes          float64   `gorm:"type:decimal(10,2);not null" json:"taxes"`
	Penalty        float64   `gorm:"type:decimal(10,2);default:0;not null" json:"penalty"`
	Discount       float64   `gorm:"type:decimal(10,2);default:0;not null" json:"discount"`
	TotalAmount    float64   `gorm:"type:decimal(10,2);not null" json:"totalAmount"`
	DriverEarnings float64   `gorm:"type:decimal(10,2);not null" json:"driverEarnings"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (Bill) TableName() string {
	return "Bills"
}
