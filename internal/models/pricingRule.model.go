package models

import (
	"time"
)

// PricingRule represents the pricing rule model
type PricingRule struct {
	ID            uint     `gorm:"primaryKey" json:"id"`
	Category      string   `gorm:"type:enum('Base','Commission','Tax','Penalty','Timings');not null" json:"category"`
	Scope         string   `gorm:"type:enum('Global','State','City');default:'Global';not null" json:"scope"`
	State         *string  `json:"state,omitempty"`
	City          *string  `json:"city,omitempty"`
	VehicleType   *string  `json:"vehicleType,omitempty"`
	BaseRate      *float64 `json:"baseRate,omitempty"`
	PerUnit       *float64 `json:"perUnit,omitempty"`
	Amount        *float64 `json:"amount,omitempty"`
	PerRides      *int     `json:"perRides,omitempty"`
	Name          *string  `json:"name,omitempty"`
	Value         *float64 `json:"value,omitempty"`
	TaxType       *string  `gorm:"type:enum('Percentage','Fixed')" json:"taxType,omitempty"`
	Role          *string  `gorm:"type:enum('Driver','Customer','CampOwner')" json:"role,omitempty"`
	CancelLimit   *int     `json:"cancelLimit,omitempty"`
	PenaltyAmount *float64 `json:"penaltyAmount,omitempty"`
	AcceptTime    *int     `json:"acceptTime,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (PricingRule) TableName() string {
	return "pricing_rules"
}
