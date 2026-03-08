package controllers

import (
	"encoding/json"	
	"fmt"
	"net/http"
	"os"
	"spark/internal/database"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"	
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func SponsorLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var sponsor models.Sponsor
	if err := database.DB.Where("username = ? AND password = ?", req.Username, req.Password).First(&sponsor).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Println("⚠️ WARNING: JWT_SECRET is not set! Using insecure default.")
		secret = "default_secret_key_change_me"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       sponsor.ID,
		"username": sponsor.Username,
		"role":     sponsor.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ Sponsor Logged In: %s (ID: %d)\n", sponsor.Username, sponsor.ID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":          tokenString,
		"user":           sponsor,
		"remainingLimit": sponsor.RemainingLimit,
	})
}

func GetSponsorHistory(w http.ResponseWriter, r *http.Request) {
	sponsor := r.Context().Value(middleware.SponsorContextKey).(*models.Sponsor)
	var history []models.SponsorNotification
	database.DB.Where("sponsor_id = ?", sponsor.ID).Order("sent_at DESC").Find(&history)
	json.NewEncoder(w).Encode(history)
}

func SendSponsorNotification(w http.ResponseWriter, r *http.Request) {
	sponsor := r.Context().Value(middleware.SponsorContextKey).(*models.Sponsor)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	if sponsor.RemainingLimit <= 0 {
		http.Error(w, "Limit reached", http.StatusForbidden)
		return
	}

	title := r.FormValue("title")
	message := r.FormValue("message")
	target := r.FormValue("target")

	// Logic to save notification and broadcast
	notification := models.SponsorNotification{
		SponsorID: sponsor.ID,
		Title:     title,
		Message:   message,
		Target:    target,
		Status:    "sending",
		SentAt:    time.Now(),
	}
	database.DB.Create(&notification)

	// Broadcast logic (mocked service call)
	counts, _ := services.BroadcastSponsorNotification(target, title, message, notification.ID)

	notification.RecipientCount = counts.DriverCount + counts.CustomerCount
	notification.DriverCount = counts.DriverCount
	notification.CustomerCount = counts.CustomerCount
	notification.Status = "sent"
	database.DB.Save(&notification)

	sponsor.RemainingLimit -= 1
	database.DB.Save(sponsor)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"notification":   notification,
		"remainingLimit": sponsor.RemainingLimit,
	})
}

func DeleteSponsorNotification(w http.ResponseWriter, r *http.Request) {
	sponsor := r.Context().Value(middleware.SponsorContextKey).(*models.Sponsor)
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)

	var notif models.SponsorNotification
	if err := database.DB.Where("id = ? AND sponsor_id = ?", id, sponsor.ID).First(&notif).Error; err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if notif.Status == "scheduled" {
		sponsor.RemainingLimit += 1
		database.DB.Save(sponsor)
	}

	database.DB.Delete(&notif)
	json.NewEncoder(w).Encode(map[string]string{"message": "Deleted"})
}

func UploadCampaignBanner(w http.ResponseWriter, r *http.Request) {
	sponsor := r.Context().Value(middleware.SponsorContextKey).(*models.Sponsor)

	// Limit upload size to 10MB
	r.ParseMultipartForm(10 << 20)

	file, header, err := r.FormFile("banner")
	if err != nil {
		http.Error(w, "No banner file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	bannerURL, err := services.UploadFile(file, header.Filename, "sponsors")
	if err != nil {
		http.Error(w, "Failed to upload banner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sponsor.BannerImage = &bannerURL
	database.DB.Save(sponsor)

	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Banner uploaded successfully", "user": sponsor})
}

func GetGamToken(w http.ResponseWriter, r *http.Request) {
	token, err := services.GetGoogleAdManagerToken()
	if err != nil {
		http.Error(w, "Failed to get GAM token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}