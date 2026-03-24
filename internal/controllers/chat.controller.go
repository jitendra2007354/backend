package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strconv"
)

func SendChatMessage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req struct {
		RideID  uint   `json:"rideId"`
		Message string `json:"message"`
		Text    string `json:"text"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	msgText := req.Message
	if msgText == "" {
		msgText = req.Text
	}

	msg, err := services.SaveChatMessage(req.RideID, user.ID, msgText)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func GetChatMessages(w http.ResponseWriter, r *http.Request) {
	rideID, _ := strconv.Atoi(r.PathValue("rideId"))
	// Optional: Check if user is part of the ride
	messages, _ := services.GetChatHistory(uint(rideID))
	json.NewEncoder(w).Encode(messages)
}

func GetUserChats(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chats, err := services.GetUserChats(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(chats)
}
