package services

import (
	"fmt"
	"spark/internal/models"
)

func SendAdminNotification(target interface{}, title, message string) error {
	var users []models.User
	query := DB.Model(&models.User{})

	switch t := target.(type) {
	case string:
		if t == "drivers" {
			query = query.Where("user_type = ?", "Driver")
		} else if t == "customers" {
			query = query.Where("user_type = ?", "Customer")
		}
	case []uint:
		query = query.Where("id IN ?", t)
	}

	if err := query.Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		DB.Create(&models.Notification{UserID: user.ID, Title: title, Message: message, Type: "general"})
		SendMessageToUser(user.ID, "admin_notification", map[string]string{"title": title, "message": message})
	}
	fmt.Printf("Sent notification to %d users\n", len(users))
	return nil
}

type BroadcastCounts struct {
	DriverCount   int
	CustomerCount int
}

func BroadcastSponsorNotification(target, title, message string, notificationID uint) (BroadcastCounts, error) {
	var users []models.User
	query := DB.Model(&models.User{})

	if target == "drivers" {
		query = query.Where("user_type = ?", "Driver")
	} else if target == "customers" {
		query = query.Where("user_type = ?", "Customer")
	}
	// If target is "all", we fetch everyone

	if err := query.Find(&users).Error; err != nil {
		return BroadcastCounts{}, err
	}

	// In a real app, you might use Firebase (FCM) here.
	// For now, we use the socket service to send to online users.
	for _, user := range users {
		SendMessageToUser(user.ID, "sponsor_notification", map[string]interface{}{
			"id":      notificationID,
			"title":   title,
			"message": message,
		})
	}

	// Calculate counts (simplified)
	drivers := 0
	customers := 0
	for _, u := range users {
		if u.UserType == "Driver" {
			drivers++
		} else {
			customers++
		}
	}

	return BroadcastCounts{DriverCount: drivers, CustomerCount: customers}, nil
}
