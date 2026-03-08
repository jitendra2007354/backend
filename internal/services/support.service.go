package services

import "spark/internal/models"

func OpenTicket(userID uint, subject, description string) (*models.SupportTicket, error) {
	ticket := models.SupportTicket{
		UserID:      userID,
		Subject:     subject,
		Description: description,
		Status:      "OPEN",
	}
	if err := DB.Create(&ticket).Error; err != nil {
		return nil, err
	}
	SendMessageToAdminRoom("new_support_ticket", ticket)
	return &ticket, nil
}

func CloseTicket(ticketID uint) error {
	DB.Model(&models.SupportTicket{}).Where("id = ?", ticketID).Update("status", "CLOSED")
	return nil
}
