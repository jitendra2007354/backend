package models

import (
	"time"

	"gorm.io/datatypes"
)

// NotificationForSponsor represents the notification for sponsor model
type NotificationForSponsor struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	SponsorID uint           `gorm:"not null" json:"sponsorId"`
	Title     string         `gorm:"not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Type      string         `gorm:"default:'info'" json:"type"`
	Read      bool           `gorm:"default:false" json:"read"`
	Likes     int            `gorm:"default:0" json:"likes"`
	Liked     bool           `gorm:"default:false" json:"liked"`
	Media     datatypes.JSON `json:"media,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func (NotificationForSponsor) TableName() string {
	return "notification_for_sponsor"
}
