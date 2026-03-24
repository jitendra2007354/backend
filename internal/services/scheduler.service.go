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
			// Try to acquire a lock in Redis for 50 seconds to prevent duplicate runs across multiple servers
			if RedisClient != nil {
				if locked, _ := RedisClient.SetNX(redisCtx, "lock:scheduler", "1", 50*time.Second).Result(); !locked {
					continue // Another server instance already claimed this minute's job! Go back to sleep.
				}
			}

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
				n.SentAt = time.Now()
				DB.Save(&n)
			}

			// Auto-cancel ghost rides that have been pending for over 30 minutes
			expirationTime := time.Now().Add(-30 * time.Minute)
			res := DB.Model(&models.Ride{}).Where("status IN ? AND created_at < ?", []string{"pending", "draft"}, expirationTime).Update("status", "cancelled")
			if res.RowsAffected > 0 {
				fmt.Printf("Auto-cancelled %d abandoned pending rides\n", res.RowsAffected)
			}
		}
	}()
}
