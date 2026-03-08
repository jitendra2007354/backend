package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/database"
	"spark/internal/models"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	// Check existence logic would go here or in DB constraint

	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
