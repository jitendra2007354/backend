package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
)

func UploadDocumentController(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	r.ParseMultipartForm(10 << 20)
	docType := r.FormValue("documentType")
	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "No file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	result, err := services.UploadDocument(user.ID, docType, header.Filename, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func VerifyDocumentController(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DriverID     uint   `json:"driverId"`
		DocumentType string `json:"documentType"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := services.VerifyDocument(req.DriverID, req.DocumentType); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
