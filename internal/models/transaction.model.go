package models

import "time"

type Transaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"userId"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`        // "credit", "debit"
	Purpose     string    `json:"purpose"`     // "wallet_topup", "fee_payment"
	ReferenceID string    `json:"referenceId"` // Stripe PaymentIntentID
	CreatedAt   time.Time `json:"createdAt"`
}
