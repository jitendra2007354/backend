package models

import (
	"time"

	"gorm.io/datatypes"
)

// SponsorNotification represents the sponsor notification model
type SponsorNotification struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	SponsorID      uint           `gorm:"not null" json:"sponsorId"`
	Sponsor        Sponsor        `gorm:"foreignKey:SponsorID" json:"sponsor,omitempty"`
	Title          string         `gorm:"not null" json:"title"`
	Message        string         `gorm:"type:text;not null" json:"message"`
	Target         string         `gorm:"not null" json:"target"`
	Attachments    datatypes.JSON `json:"attachments,omitempty"`
	SentAt         time.Time      `json:"sentAt"`
	ScheduledFor   *time.Time     `json:"scheduledFor,omitempty"`
	Status         string         `gorm:"default:'sent'" json:"status"`
	RecipientCount int            `gorm:"default:0" json:"recipientCount"`
	DriverCount    int            `gorm:"default:0" json:"driverCount"`
	CustomerCount  int            `gorm:"default:0" json:"customerCount"`
}

func (SponsorNotification) TableName() string {
	return "SponsorNotifications"
}