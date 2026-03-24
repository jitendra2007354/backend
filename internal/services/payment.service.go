package services

import (
	"errors"
	"fmt"
	"math"
	"os"
	"spark/internal/models"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"gorm.io/gorm"
)

type PaymentResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	TransactionID string `json:"transactionId,omitempty"`
	ClientSecret  string `json:"clientSecret,omitempty"`
}

func ProcessPayment(userID uint, amount float64, purpose string) (*PaymentResult, error) {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		return FulfillPayment(userID, amount, purpose, fmt.Sprintf("mock_%d", time.Now().UnixNano()))
	}

	stripe.Key = stripeKey

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)), // Convert to cents
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: map[string]string{
			"user_id": fmt.Sprintf("%d", userID),
			"purpose": purpose,
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return &PaymentResult{Success: false, Message: err.Error()}, err
	}

	// Return the client secret so the frontend can confirm the payment
	return &PaymentResult{
		Success:       true,
		Message:       "Payment initiated",
		TransactionID: pi.ID,
		ClientSecret:  pi.ClientSecret,
	}, nil
}

func ConfirmPayment(paymentIntentID string) (*PaymentResult, error) {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		return nil, errors.New("stripe not configured")
	}
	stripe.Key = stripeKey

	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}

	if pi.Status != stripe.PaymentIntentStatusSucceeded {
		return &PaymentResult{Success: false, Message: "Payment not successful: " + string(pi.Status)}, nil
	}

	// Extract User ID from metadata
	var userID uint
	fmt.Sscanf(pi.Metadata["user_id"], "%d", &userID)

	purpose := pi.Metadata["purpose"]
	if purpose == "" {
		purpose = "fee_payment" // Default fallback
	}

	// Credit the user (using the mock logic which handles DB updates)
	amount := float64(pi.Amount) / 100.0
	return FulfillPayment(userID, amount, purpose, pi.ID)
}

func FulfillPayment(userID uint, amount float64, purpose string, referenceID string) (*PaymentResult, error) {
	// 1. Idempotency Check: Prevent double crediting
	var existingTx models.Transaction
	if err := DB.Where("reference_id = ?", referenceID).First(&existingTx).Error; err == nil {
		return &PaymentResult{Success: true, Message: "Payment already processed", TransactionID: referenceID}, nil
	}

	// 2. Atomic Transaction: Record history + Update Balance
	err := DB.Transaction(func(tx *gorm.DB) error {
		// Record the transaction
		newTx := models.Transaction{
			UserID:      userID,
			Amount:      amount,
			Type:        "credit",
			Purpose:     purpose,
			ReferenceID: referenceID,
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(&newTx).Error; err != nil {
			return err
		}

		// Update Balance based on purpose
		if purpose == "wallet_topup" {
			if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("wallet_balance", gorm.Expr("wallet_balance + ?", amount)).Error; err != nil {
				return err
			}
		} else {
			// Default: fee_payment
			var driver models.Driver
			if err := tx.Where("user_id = ?", userID).First(&driver).Error; err != nil {
				return errors.New("driver not found")
			}

			// Reduce outstanding fee, ensure it doesn't go below 0 (logic simplified)
			newBalance := math.Max(0, driver.OutstandingPlatformFee-amount)

			updates := map[string]interface{}{
				"outstanding_platform_fee": newBalance,
			}
			if err := tx.Model(&driver).Updates(updates).Error; err != nil {
				return err
			}
			if newBalance <= 0 {
				tx.Model(&driver).Update("platform_access_expiry", gorm.Expr("NULL"))
				tx.Model(&models.User{}).Where("id = ?", userID).Update("is_blocked", false)
			}
		}
		return nil
	})

	if err != nil {
		return &PaymentResult{Success: false, Message: err.Error()}, nil
	}

	return &PaymentResult{Success: true, Message: "Payment processed successfully"}, nil
}

// ProcessDailyFee charges a fixed daily fee if 24 hours have passed since last charge.
func ProcessDailyFee(driverID uint, tx *gorm.DB) {
	var driver models.Driver
	if err := tx.First(&driver, driverID).Error; err != nil {
		return
	}

	now := time.Now()
	// Only apply a new fee and start a new timer if they have cleared their previous dues.
	// This prevents the timer from restarting if it expires while they are mid-ride.
	if driver.OutstandingPlatformFee <= 0 {
		if driver.LastDailyFeeChargedAt == nil || now.Sub(*driver.LastDailyFeeChargedAt) > 1*time.Minute {
			dailyFee := 50.0

			// Dynamically fetch platform_fee using SQL select to cleanly auto-convert numeric types
			var dbFee float64
			if err := tx.Table("configs").Select("platform_fee").Where("`key` = ?", "global").Scan(&dbFee).Error; err == nil && dbFee > 0 {
				dailyFee = dbFee
			}
			// update fields
			driver.OutstandingPlatformFee += dailyFee
			driver.LastDailyFeeChargedAt = &now

			// Set expiry to 1 minute
			expiry := now.Add(1 * time.Minute)
			driver.PlatformAccessExpiry = &expiry
			tx.Save(&driver)
		}
	}
}
