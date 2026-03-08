package services

import "spark/internal/models"

func BuildSponsorPayload(sponsor *models.Sponsor) map[string]interface{} {
	return map[string]interface{}{
		"username":           sponsor.Username,
		"role":               sponsor.Role,
		"remainingLimit":     sponsor.RemainingLimit,
		"totalLimit":         sponsor.TotalLimit,
		"validUntil":         sponsor.ValidUntil,
		"profileImage":       sponsor.ProfileImage,
		"customHtmlTemplate": sponsor.CustomHTMLTemplate,
		"gamReportsEnabled":  sponsor.GAMReportsEnabled,
	}
}