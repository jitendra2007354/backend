package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
)

func SubmitSupportTicket(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		Subject     string `json:"subject"`
		Message     string `json:"message"`
		Description string `json:"description"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	body := req.Description
	if body == "" {
		body = req.Message
	}

	ticket, err := services.OpenTicket(user.ID, req.Subject, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ticket)
}
