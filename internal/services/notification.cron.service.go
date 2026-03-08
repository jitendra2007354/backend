package services

import (
	"fmt"
	"time"
)

func ScheduleSponsorNotificationCleanup() {
	ticker := time.NewTicker(24 * time.Hour * 31) // Approx monthly
	go func() {
		for range ticker.C {
			fmt.Println("Cleaning up sponsor notifications...")
			// DB delete logic
		}
	}()
}