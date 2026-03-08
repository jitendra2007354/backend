package models

import (
	"time"
)

// Rating represents the rating model
type Rating struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RideID    uint      `gorm:"unique;not null" json:"rideId"`
	Ride      Ride      `gorm:"foreignKey:RideID" json:"ride,omitempty"`
	RaterID   uint      `gorm:"not null" json:"raterId"`
	Rater     User      `gorm:"foreignKey:RaterID" json:"rater,omitempty"`
	RatedID   uint      `gorm:"not null" json:"ratedId"`
	Rated     User      `gorm:"foreignKey:RatedID" json:"rated,omitempty"`
	Rating    int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   *string   `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Rating) TableName() string {
	return "ratings"
}
