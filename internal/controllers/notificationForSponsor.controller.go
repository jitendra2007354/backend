package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/database"
	"spark/internal/middleware"
	"spark/internal/models"
	"strconv"
)

func GetSystemNotifications(w http.ResponseWriter, r *http.Request) {
	sponsor := r.Context().Value(middleware.SponsorContextKey).(*models.Sponsor)
	var notifs []models.NotificationForSponsor
	database.DB.Where("sponsor_id = ?", sponsor.ID).Order("created_at DESC").Find(&notifs)
	json.NewEncoder(w).Encode(notifs)
}

func MarkSystemNotificationRead(w http.ResponseWriter, r *http.Request) {
	sponsor := r.Context().Value(middleware.SponsorContextKey).(*models.Sponsor)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	database.DB.Model(&models.NotificationForSponsor{}).Where("id = ? AND sponsor_id = ?", id, sponsor.ID).Update("read", true)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func MarkAllSystemNotificationsRead(w http.ResponseWriter, r *http.Request) {
	sponsor := r.Context().Value(middleware.SponsorContextKey).(*models.Sponsor)
	database.DB.Model(&models.NotificationForSponsor{}).Where("sponsor_id = ? AND read = ?", sponsor.ID, false).Update("read", true)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func ToggleSystemNotificationLike(w http.ResponseWriter, r *http.Request) {
	sponsor := r.Context().Value(middleware.SponsorContextKey).(*models.Sponsor)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var req struct {
		Liked bool `json:"liked"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var notif models.NotificationForSponsor
	if err := database.DB.Where("id = ? AND sponsor_id = ?", id, sponsor.ID).First(&notif).Error; err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	increment := 1
	if !req.Liked {
		increment = -1
	}
	notif.Liked = req.Liked
	notif.Likes += increment
	if notif.Likes < 0 {
		notif.Likes = 0
	}
	database.DB.Save(&notif)

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "likes": notif.Likes})
}
