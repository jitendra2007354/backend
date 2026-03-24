package services

import (
	"fmt"
	"spark/internal/models"
)

func SaveChatMessage(rideID, senderID uint, message string) (*models.ChatMessage, error) {
	var ride models.Ride
	if err := DB.First(&ride, rideID).Error; err != nil {
		return nil, err
	}

	receiverID := ride.CustomerID
	if senderID == ride.CustomerID {
		if ride.DriverID != nil {
			var driver models.Driver
			if err := DB.First(&driver, *ride.DriverID).Error; err == nil {
				receiverID = driver.UserID
			}
		}
	}

	msg := models.ChatMessage{
		RideID:     rideID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Message:    message,
	}
	err := DB.Create(&msg).Error
	return &msg, err
}

type FrontendMessage struct {
	ID       uint   `json:"id"`
	Text     string `json:"text"`
	IsSender bool   `json:"isSender"`
	Time     string `json:"time"`
}

type FrontendChat struct {
	ID                 uint              `json:"id"`
	DriverName         string            `json:"driverName"`
	DriverMobileNumber string            `json:"driverMobileNumber"`
	AvatarInitial      string            `json:"avatarInitial"`
	Status             string            `json:"status"`
	Timestamp          string            `json:"timestamp"`
	UnreadCount        int               `json:"unreadCount"`
	Messages           []FrontendMessage `json:"messages"`
}

func GetUserChats(userID uint) ([]FrontendChat, error) {
	var msgs []models.ChatMessage
	if err := DB.Where("sender_id = ? OR receiver_id = ?", userID, userID).Order("created_at ASC").Find(&msgs).Error; err != nil {
		return nil, err
	}

	chatMap := make(map[uint]*FrontendChat)
	rideIds := make([]uint, 0)

	for _, m := range msgs {
		if _, exists := chatMap[m.RideID]; !exists {
			chatMap[m.RideID] = &FrontendChat{
				ID:       m.RideID,
				Messages: []FrontendMessage{},
			}
			rideIds = append(rideIds, m.RideID)
		}

		isSender := m.SenderID == userID
		chatMap[m.RideID].Messages = append(chatMap[m.RideID].Messages, FrontendMessage{
			ID:       m.ID,
			Text:     m.Message,
			IsSender: isSender,
			Time:     m.CreatedAt.Format("15:04"),
		})
		chatMap[m.RideID].Timestamp = m.CreatedAt.Format("15:04")
	}

	var rides []models.Ride
	if len(rideIds) > 0 {
		DB.Preload("Customer").Preload("Driver.User").Where("id IN ?", rideIds).Find(&rides)
		for _, r := range rides {
			chat, ok := chatMap[r.ID]
			if !ok {
				continue
			}
			if r.CustomerID == userID {
				if r.Driver != nil {
					chat.DriverName = fmt.Sprintf("%s %s", r.Driver.User.FirstName, r.Driver.User.LastName)
					chat.DriverMobileNumber = r.Driver.User.PhoneNumber
					if len(chat.DriverName) > 0 {
						chat.AvatarInitial = string(chat.DriverName[0])
					}
				}
			} else {
				chat.DriverName = fmt.Sprintf("%s %s", r.Customer.FirstName, r.Customer.LastName)
				chat.DriverMobileNumber = r.Customer.PhoneNumber
				if len(chat.DriverName) > 0 {
					chat.AvatarInitial = string(chat.DriverName[0])
				}
			}
			chat.Status = "Online"
		}
	}

	var result []FrontendChat
	for _, id := range rideIds {
		result = append(result, *chatMap[id])
	}

	// Reverse to get latest chats first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	if result == nil {
		result = []FrontendChat{}
	}

	return result, nil
}
func GetChatHistory(rideID uint) ([]models.ChatMessage, error) {
	var msgs []models.ChatMessage
	err := DB.Where("ride_id = ?", rideID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}
