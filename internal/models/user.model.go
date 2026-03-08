package models

import (
	"time"
)

// User represents the user model
type User struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	FirstName              string     `gorm:"not null" json:"firstName"`
	LastName               string     `gorm:"not null" json:"lastName"`
	Email                  *string    `gorm:"unique" json:"email,omitempty"`
	PhoneNumber            string     `gorm:"unique;not null" json:"phoneNumber"`
	PFP                    *string    `json:"pfp,omitempty"`
	City                   *string    `json:"city,omitempty"`
	State                  *string    `json:"state,omitempty"`
	UserType               string     `gorm:"type:enum('Customer','Driver','Admin');default:'Customer'" json:"userType"`
	IsOnline               bool       `gorm:"default:false" json:"isOnline"`
	IsBlocked              bool       `gorm:"default:false" json:"isBlocked"`
	WalletBalance          float64    `gorm:"default:0" json:"walletBalance"`
	LowBalanceSince        *time.Time `json:"lowBalanceSince,omitempty"`
	DriverPicURL           *string    `json:"driverPicUrl,omitempty"`
	LicenseURL             *string    `json:"licenseUrl,omitempty"`
	RCURL                  *string    `json:"rcUrl,omitempty"`
	DriverPicIsVerified    bool       `gorm:"default:false" json:"driverPicIsVerified"`
	LicenseIsVerified      bool       `gorm:"default:false" json:"licenseIsVerified"`
	RCIsVerified           bool       `gorm:"default:false" json:"rcIsVerified"`
	AverageRating          float64    `gorm:"default:0" json:"averageRating"`
	OutstandingPlatformFee float64    `gorm:"default:0" json:"outstandingPlatformFee"`
	CurrentLat             *float64   `json:"currentLat,omitempty"`
	CurrentLng             *float64   `json:"currentLng,omitempty"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}

func (User) TableName() string {
	return "Users"
}