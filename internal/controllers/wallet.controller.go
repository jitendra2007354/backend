package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
)

func GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	balance, err := services.GetDriverWalletBalance(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
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
	json.NewDecoder(r.Body).Decode(&req)

	updatedDriver, _ := services.UpdateUserWallet(req.DriverID, req.Amount)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Adjusted", "newBalance": updatedDriver.WalletBalance, "isBlocked": updatedDriver.IsBlocked,
	})
}
