package services

import "spark/internal/models"

func SaveChatMessage(rideID, senderID uint, message string) (*models.ChatMessage, error) {
	// Determine receiver logic omitted for brevity, assuming passed or derived
	receiverID := uint(0) // Placeholder

	msg := models.ChatMessage{
		RideID:     rideID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Message:    message,
	}
	err := DB.Create(&msg).Error
	return &msg, err
}

func GetChatHistory(rideID uint) ([]models.ChatMessage, error) {
	var msgs []models.ChatMessage
	err := DB.Where("ride_id = ?", rideID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}
