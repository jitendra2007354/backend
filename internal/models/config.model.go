package models

import (
	"time"
)

// Config represents the configuration model
type Config struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	Key                     string    `gorm:"unique;not null" json:"key"`
	BaseFare                float64   `gorm:"type:decimal(10,2);not null" json:"baseFare"`
	BaseFareBike            float64   `gorm:"type:decimal(10,2);default:25;not null" json:"baseFareBike"`
	BaseFareAuto            float64   `gorm:"type:decimal(10,2);default:35;not null" json:"baseFareAuto"`
	BaseFareCar             float64   `gorm:"type:decimal(10,2);default:50;not null" json:"baseFareCar"`
	BaseFareSUV             float64   `gorm:"type:decimal(10,2);default:65;not null" json:"baseFareSuv"`
	BaseFareLuxury          float64   `gorm:"type:decimal(10,2);default:100;not null" json:"baseFareLuxury"`
	PerKmRate               float64   `gorm:"type:decimal(10,2);not null" json:"perKmRate"`
	PerMinuteRate           float64   `gorm:"type:decimal(10,2);default:1;not null" json:"perMinuteRate"`
	CommissionRate          float64   `gorm:"type:decimal(5,2);not null" json:"commissionRate"`
	CancellationFee         float64   `gorm:"type:decimal(10,2);not null" json:"cancellationFee"`
	DriverSearchRadius      float64   `gorm:"type:decimal(10,2);not null" json:"driverSearchRadius"`
	RideAcceptTime          int       `gorm:"not null" json:"rideAcceptTime"`
	WalletMinBalance        float64   `gorm:"type:decimal(10,2);not null" json:"walletMinBalance"`
	MaxBidPerKm             float64   `gorm:"type:decimal(10,2);default:30;not null" json:"maxBidPerKm"`
	MinBidPerKm             float64   `gorm:"type:decimal(10,2);default:8;not null" json:"minBidPerKm"`
	MinBidPerKmBike         float64   `gorm:"type:decimal(10,2);default:4.8;not null" json:"minBidPerKmBike"`
	MinBidPerKmAuto         float64   `gorm:"type:decimal(10,2);default:6.4;not null" json:"minBidPerKmAuto"`
	MinBidPerKmCar          float64   `gorm:"type:decimal(10,2);default:8;not null" json:"minBidPerKmCar"`
	MinBidPerKmSUV          float64   `gorm:"type:decimal(10,2);default:9.6;not null" json:"minBidPerKmSuv"`
	MinBidPerKmLuxury       float64   `gorm:"type:decimal(10,2);default:14.4;not null" json:"minBidPerKmLuxury"`
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
