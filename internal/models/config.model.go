package models

import (
	"time"
)

// Config represents the configuration model
type Config struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	Key                     string    `gorm:"unique;not null" json:"key"`
	BaseFare                float64   `gorm:"type:decimal(10,2);not null" json:"baseFare"`
	PerKmRate               float64   `gorm:"type:decimal(10,2);not null" json:"perKmRate"`
	PerMinuteRate           float64   `gorm:"type:decimal(10,2);default:1;not null" json:"perMinuteRate"`
	CommissionRate          float64   `gorm:"type:decimal(5,2);not null" json:"commissionRate"`
	CancellationFee         float64   `gorm:"type:decimal(10,2);not null" json:"cancellationFee"`
	DriverSearchRadius      float64   `gorm:"type:decimal(10,2);not null" json:"driverSearchRadius"`
	RideAcceptTime          int       `gorm:"not null" json:"rideAcceptTime"`
	WalletMinBalance        float64   `gorm:"type:decimal(10,2);not null" json:"walletMinBalance"`
	AutoBlockHours          int       `gorm:"default:24;not null" json:"autoBlockHours"`
	SurgeMultiplier         float64   `gorm:"type:decimal(5,2);default:1;not null" json:"surgeMultiplier"`
	TaxRate                 float64   `gorm:"type:decimal(5,2);default:5;not null" json:"taxRate"`
	CancellationGracePeriod int       `gorm:"default:60;not null" json:"cancellationGracePeriod"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

func (Config) TableName() string {
	return "configs"
}
