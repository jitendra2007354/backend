package models

import (
	"time"

	"gorm.io/datatypes"
)

// Sponsor represents the sponsor model
type Sponsor struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	Username           string         `gorm:"unique;not null" json:"username"`
	Password           string         `gorm:"not null" json:"-"`
	Name               *string        `json:"name,omitempty"`
	ProfileImage       *string        `json:"profileImage,omitempty"`
	Role               string         `gorm:"type:enum('admin','editor');default:'editor';not null" json:"role"`
	RemainingLimit     int            `gorm:"default:100;not null" json:"remainingLimit"`
	TotalLimit         int            `gorm:"default:100;not null" json:"totalLimit"`
	ValidUntil         *time.Time     `json:"validUntil,omitempty"`
	CustomHTMLTemplate bool           `gorm:"default:false;not null" json:"customHtmlTemplate"`
	GAMReportsEnabled  bool           `gorm:"default:false;not null" json:"gamReportsEnabled"`
	GAMAdvertiserID    *string        `json:"gamAdvertiserId,omitempty"`
	GAMOrderID         *string        `json:"gamOrderId,omitempty"`
	GAMLineItemID      *string        `json:"gamLineItemId,omitempty"`
	GAMNetworkCode     *string        `json:"gamNetworkCode,omitempty"`
	GAMAdUnitID        *string        `json:"gamAdUnitId,omitempty"`
	ServiceAccount     datatypes.JSON `json:"serviceAccount,omitempty"`
	BannerImage        *string        `json:"bannerImage,omitempty"`
}

// TableName overrides the table name used by User to `Sponsors`
func (Sponsor) TableName() string {
	return "Sponsors"
}