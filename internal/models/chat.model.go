package models

import (
	"time"
)

// Chat represents the chat model (distinct from ChatMessage by table name)
type Chat struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	RideID      uint      `gorm:"not null" json:"rideId"`
	Ride        Ride      `gorm:"foreignKey:RideID" json:"ride,omitempty"`
	SenderID    uint      `gorm:"not null" json:"senderId"`
	Sender      User      `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReceiverID  uint      `gorm:"not null" json:"receiverId"`
	Receiver    User      `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	Message     *string   `gorm:"type:text" json:"message,omitempty"`
	FileContent *string   `gorm:"type:longtext" json:"fileContent,omitempty"`
	FileType    *string   `json:"fileType,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (Chat) TableName() string {
	return "chat_messages"
}
