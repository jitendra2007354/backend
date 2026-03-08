package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
)

func MakePayment(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		Amount float64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := services.ProcessPayment(user.ID, req.Amount, "fee_payment")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PaymentIntentID string `json:"paymentIntentId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := services.ConfirmPayment(req.PaymentIntentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}
