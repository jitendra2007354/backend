package services

import (
	"fmt"
	"spark/internal/models"
	"time"
)

// ScheduleChatCleanup deletes chat messages belonging to rides that
// were completed or cancelled.
func ScheduleChatCleanup() {
	go func() {
		for {
			now := time.Now()
			nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := nextMidnight.Sub(now)
			time.Sleep(duration)

			// Try to acquire a lock in Redis for 10 minutes.
			// Since all servers wake up at exact midnight, the first one to hit Redis wins.
			if RedisClient != nil {
				if locked, _ := RedisClient.SetNX(redisCtx, "lock:chat_cleanup", "1", 10*time.Minute).Result(); !locked {
					continue
				}
			}

			fmt.Println("Running scheduled job: deleting completed/cancelled chat messages...")
			var oldRides []models.Ride
			if err := DB.Select("id").Where("status IN ?", []string{"completed", "cancelled"}).Find(&oldRides).Error; err != nil {
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
