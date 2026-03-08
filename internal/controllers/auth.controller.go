package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/services"
)

func Login(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form for file uploads
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Helper to save file and return URL
	saveFile := func(field string) string {
		file, header, err := r.FormFile(field)
		if err != nil {
			return ""
		}
		defer file.Close()
		
		url, _ := services.UploadFile(file, header.Filename, "users")
		return url
	}

	// Helper to get form value
	get := func(key string) string { return r.FormValue(key) }

	loginData := map[string]interface{}{
		"phoneNumber":           get("phoneNumber"),
		"firstName":             get("firstName"),
		"lastName":              get("lastName"),
		"email":                 get("email"),
		"city":                  get("city"),
		"state":                 get("state"),
		"vehicleType":           get("vehicleType"),
		"vehicleModel":          get("vehicleModel"),
		"vehicleNumber":         get("vehicleNumber"),
		"driverLicenseNumber":   get("driverLicenseNumber"),
		"driverLicensePhotoUrl": saveFile("driverLicensePhotoUrl"),
		"rcPhotoUrl":            saveFile("rcPhotoUrl"),
		"pfp":                   saveFile("pfp"),
	}

	result, err := services.LoginOrRegister(loginData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   result.Token,
		"user":    result.User,
	})
}

func SendOtp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mobile string `json:"mobile"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := services.SendOtpService(req.Mobile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "OTP sent"})
}

func VerifyOtp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mobile string                 `json:"mobile"`
		Otp    string                 `json:"otp"`
		Data   map[string]interface{} `json:"data"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := services.VerifyOtpService(req.Mobile, req.Otp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Construct login data from req.Data + Mobile
	loginData := req.Data
	loginData["phoneNumber"] = req.Mobile

	result, err := services.LoginOrRegister(loginData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   result.Token,
		"user":    result.User,
	})
}

func GuestLogin(w http.ResponseWriter, r *http.Request) {
	result, err := services.LoginGuest()
	if err != nil {
		http.Error(w, "Failed to login as guest", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Guest login successful",
		"token":   result.Token,
		"user":    result.User,
	})
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := services.LoginAdmin(req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func Me(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey)
	json.NewEncoder(w).Encode(user)
}
