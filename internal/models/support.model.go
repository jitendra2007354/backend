package models

import (
	"time"
)

// SupportTicket represents the support ticket model
type SupportTicket struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"userId"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Subject     string    `gorm:"not null" json:"subject"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Status      string    `gorm:"type:enum('OPEN','IN_PROGRESS','RESOLVED','CLOSED');default:'OPEN'" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (SupportTicket) TableName() string {
	return "support_tickets"
}
