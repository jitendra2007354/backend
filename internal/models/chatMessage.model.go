package models

import (
	"time"
)

// ChatMessage represents the chat message model
type ChatMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RideID     uint      `gorm:"not null" json:"rideId"`
	Ride       Ride      `gorm:"foreignKey:RideID" json:"ride,omitempty"`
	SenderID   uint      `gorm:"not null" json:"senderId"`
	Sender     User      `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReceiverID uint      `gorm:"not null" json:"receiverId"`
	Receiver   User      `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	Message    string    `gorm:"type:text;not null" json:"message"`
	IsRead     bool      `gorm:"default:false;not null" json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (ChatMessage) TableName() string {
	return "ChatMessages"
}
