package services

import (
	"fmt"
	"time"
)

func ScheduleSponsorNotificationCleanup() {
	ticker := time.NewTicker(24 * time.Hour * 31) // Approx monthly
	go func() {
		for range ticker.C {
			if RedisClient != nil {
				if locked, _ := RedisClient.SetNX(redisCtx, "lock:sponsor_cleanup", "1", 24*time.Hour).Result(); !locked {
					continue
				}
			}

			fmt.Println("Cleaning up sponsor notifications...")
			// DB delete logic
		}
	}()
}
