package services

import (
	"fmt"
	"time"
	"spark/internal/models"
)

// ScheduleChatCleanup deletes chat messages belonging to rides that
// were completed or cancelled more than 24 hours ago.
func ScheduleChatCleanup() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			fmt.Println("Running scheduled job: deleting old chat messages...")
			cutoff := time.Now().Add(-24 * time.Hour)
			var oldRides []models.Ride
			if err := DB.Select("id").Where("status IN ? AND updated_at < ?", []string{"completed", "cancelled"}, cutoff).Find(&oldRides).Error; err != nil {
				fmt.Println("Chat cleanup fetch error:", err)
				continue
			}
			if len(oldRides) == 0 {
				fmt.Println("No rides eligible for chat cleanup")
				continue
			}
			rideIds := make([]uint, len(oldRides))
			for i, r := range oldRides {
				rideIds[i] = r.ID
			}
			res := DB.Where("ride_id IN ?", rideIds).Delete(&models.ChatMessage{})
			fmt.Printf("Deleted %d chat messages from %d rides\n", res.RowsAffected, len(rideIds))
		}
	}()
}

// ScheduleDriverStatusChecks blocks drivers who haven't paid their outstanding
// fee within 24 hours of it being charged.
func ScheduleDriverStatusChecks() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			fmt.Println("Running scheduled job: checking for drivers with outstanding balances...")
			paymentDeadline := time.Now().Add(-24 * time.Hour)
			var drivers []models.Driver
			if err := DB.Preload("User").Where("outstanding_platform_fee > ?", 0).Find(&drivers).Error; err != nil {
				fmt.Println("Driver status check query error:", err)
				continue
			}
			for _, drv := range drivers {
				if drv.User.IsBlocked {
					continue
				}
				if drv.LastDailyFeeChargedAt != nil && drv.LastDailyFeeChargedAt.Before(paymentDeadline) {
					// block the associated user
					DB.Model(&models.User{}).Where("id = ?", drv.UserID).Update("is_blocked", true)
					fmt.Printf("Auto-blocked driver %d (user %d) for non-payment\n", drv.ID, drv.UserID)
					SendMessageToUser(drv.UserID, "account_blocked", map[string]interface{}{"reason": "Payment overdue: Daily access fee not paid within 24 hours."})
				}
			}
		}
	}()
}