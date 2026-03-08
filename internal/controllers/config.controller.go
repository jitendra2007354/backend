package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/services"
)

func UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	if err := services.SetConfig(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Updated"})
}

func GetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := services.GetApplicableConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(config)
}
