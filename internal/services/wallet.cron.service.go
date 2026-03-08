package services

import (
	"fmt"
	"time"
	"spark/internal/models"
)

const (
	minimumBalance   = 500.0
	gracePeriodHours = 72
)

// ScheduleWalletCheck inspects driver wallets every hour and blocks/flags as needed.
func ScheduleWalletCheck() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			fmt.Println("Running hourly wallet balance check...")
			now := time.Now()
			var drivers []models.User
			if err := DB.Where("user_type = ? AND wallet_balance < ? AND is_blocked = false", "Driver", minimumBalance).Find(&drivers).Error; err != nil {
				fmt.Println("Wallet check query error:", err)
				continue
			}
			for _, d := range drivers {
				if d.LowBalanceSince != nil {
					expiry := d.LowBalanceSince.Add(gracePeriodHours * time.Hour)
					if now.After(expiry) {
						d.IsBlocked = true
						DB.Save(&d)
						fmt.Printf("Driver %d blocked due to prolonged low balance\n", d.ID)
					}
				} else {
					d.LowBalanceSince = &now
					DB.Save(&d)
					fmt.Printf("Driver %d entered low balance grace period\n", d.ID)
				}
			}
			// reset flag for drivers who recharged
			DB.Model(&models.User{}).Where("user_type = ? AND wallet_balance >= ? AND low_balance_since IS NOT NULL", "Driver", minimumBalance).Update("low_balance_since", nil)
		}
	}()
}
