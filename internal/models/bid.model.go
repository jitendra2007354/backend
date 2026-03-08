package models

import (
	"time"
)

// Bid represents the bid model
type Bid struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RideID     uint      `gorm:"not null;uniqueIndex:idx_ride_driver" json:"rideId"`
	Ride       Ride      `gorm:"foreignKey:RideID" json:"ride,omitempty"`
	DriverID   uint      `gorm:"not null;uniqueIndex:idx_ride_driver" json:"driverId"`
	Driver     Driver    `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Amount     float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	IsAccepted bool      `gorm:"default:false;not null" json:"isAccepted"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Bid) TableName() string {
	return "Bids"
}
