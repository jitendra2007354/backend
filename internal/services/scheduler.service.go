package services

import (
	"fmt"
	"time"

	"spark/internal/models"
)

func InitScheduler() {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			fmt.Println("Checking scheduled notifications...")
			var notifs []models.SponsorNotification
			if err := DB.Where("status = ? AND scheduled_for <= ?", "scheduled", time.Now()).Find(&notifs).Error; err != nil {
				fmt.Println("Scheduler query error:", err)
				continue
			}
			for _, n := range notifs {
				counts, _ := BroadcastSponsorNotification(n.Target, n.Title, n.Message, n.ID)
				n.RecipientCount = counts.DriverCount + counts.CustomerCount
				n.DriverCount = counts.DriverCount
				n.CustomerCount = counts.CustomerCount
				n.Status = "sent"
				DB.Save(&n)
				var sponsor models.Sponsor
				if err := DB.First(&sponsor, n.SponsorID).Error; err == nil {
					sponsor.RemainingLimit -= 1
					DB.Save(&sponsor)
				}
			}
		}
	}()
}