package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/services"
)

type NotifReq struct {
	UserID   uint   `json:"userId"`
	UserType string `json:"userType"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

func SendToUser(w http.ResponseWriter, r *http.Request) {
	var req NotifReq
	json.NewDecoder(r.Body).Decode(&req)

	// Assuming service handles single user ID logic
	services.SendAdminNotification([]uint{req.UserID}, req.Title, req.Message)
	json.NewEncoder(w).Encode(map[string]string{"message": "Sent"})
}

func SendToGroup(w http.ResponseWriter, r *http.Request) {
	var req NotifReq
	json.NewDecoder(r.Body).Decode(&req)

	services.SendAdminNotification(req.UserType, req.Title, req.Message)
	json.NewEncoder(w).Encode(map[string]string{"message": "Sent to group"})
}

func SendToAllUsers(w http.ResponseWriter, r *http.Request) {
	var req NotifReq
	json.NewDecoder(r.Body).Decode(&req)

	services.SendAdminNotification("all", req.Title, req.Message)
	json.NewEncoder(w).Encode(map[string]string{"message": "Sent to all"})
}
