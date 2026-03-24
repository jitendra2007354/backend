package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"time"
)

func GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	contextUser := r.Context().Value(middleware.UserContextKey).(*models.User)
	balance, err := services.GetDriverWalletBalance(contextUser.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var dbUser models.User
	if err := services.DB.First(&dbUser, contextUser.ID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var driver models.Driver
	outstanding := 0.0
	var expiry *time.Time
	isBlocked := dbUser.IsBlocked

	if err := services.DB.Where("user_id = ?", dbUser.ID).First(&driver).Error; err == nil {
		outstanding = driver.OutstandingPlatformFee
		expiry = driver.PlatformAccessExpiry

		// Remove backend auto-blocking based on timer. Frontend will calculate and block itself.
		if outstanding <= 0 && dbUser.IsBlocked {
			// Unblock in DB if fee is cleared
			services.DB.Model(&models.User{}).Where("id = ?", dbUser.ID).Update("is_blocked", false)
			isBlocked = false
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"balance": balance, "outstandingPlatformFee": outstanding, "platformAccessExpiry": expiry, "isBlocked": isBlocked,
	})
}

func GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	var transactions []models.Transaction
	// Fetch from the database where user_id matches, ordered by newest first
	if err := services.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&transactions).Error; err != nil {
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transactions)
}

func TopUpWallet(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		Amount float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := services.ProcessPayment(user.ID, req.Amount, "wallet_topup")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func AdjustWallet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DriverID uint    `json:"driverId"`
		Amount   float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedDriver, err := services.UpdateUserWallet(req.DriverID, req.Amount)
	if err != nil {
		http.Error(w, "Failed to update wallet: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Adjusted", "newBalance": updatedDriver.WalletBalance, "isBlocked": false,
	})
}
