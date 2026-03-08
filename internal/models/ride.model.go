package models

import (
	"time"

	"gorm.io/datatypes"
)

// Ride represents the ride model
type Ride struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	CustomerID          uint           `gorm:"not null" json:"customerId"`
	Customer            User           `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	DriverID            *uint          `json:"driverId,omitempty"`
	Driver              *Driver        `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Status              string         `gorm:"type:enum('pending','assigning','accepted','arrived','in-progress','completed','cancelled');default:'pending';not null" json:"status"`
	Fare                float64        `gorm:"type:decimal(10,2);not null" json:"fare"`
	DriverEarning       float64        `gorm:"type:decimal(10,2);default:0.00;not null" json:"driverEarning"`
	PickupLocation      datatypes.JSON `gorm:"not null" json:"pickupLocation"`
	DestinationLocation datatypes.JSON `gorm:"not null" json:"destinationLocation"`
	VehicleType         string         `gorm:"not null" json:"vehicleType"`
	Distance            *float64       `json:"distance,omitempty"`
	Duration            *float64       `json:"duration,omitempty"`
	
	// Assignment logic
	CurrentDriverID   *uint          `json:"currentDriverId,omitempty"`
	CurrentDriver     *Driver        `gorm:"foreignKey:CurrentDriverID" json:"-"`
	OfferExpiresAt    *time.Time     `json:"offerExpiresAt,omitempty"`
	RejectedDriverIDs datatypes.JSON `json:"rejectedDriverIds,omitempty"`
	
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Ride) TableName() string {
	return "rides"
}
