package models

import (
	"time"

	"gorm.io/datatypes"
)

// Notification represents the notification model
type Notification struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null" json:"userId"`
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Title       string         `gorm:"not null" json:"title"`
	Message     string         `gorm:"type:text;not null" json:"message"`
	IsRead      bool           `gorm:"default:false;not null" json:"isRead"`
	Type        string         `gorm:"type:enum('ride','offer','profile','payment','chat','general');default:'general';not null" json:"type"`
	RelatedData datatypes.JSON `json:"relatedData,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

func (Notification) TableName() string {
	return "Notifications"
}
