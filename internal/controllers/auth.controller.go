package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strings"
)

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	loginData := map[string]interface{}{}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
			http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Handle form-data fallback for file uploads and legacy clients
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
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

		phone := get("phoneNumber")
		if phone == "" {
			phone = get("mobile")
		}
		if phone == "" {
			phone = get("mobileNumber")
		}

		loginData = map[string]interface{}{
			"phoneNumber":           phone,
			"firstName":             get("firstName"),
			"lastName":              get("lastName"),
			"email":                 get("email"),
			"city":                  get("city"),
			"state":                 get("state"),
			"vehicleType":           get("vehicleType"),
			"vehicleModel":          get("vehicleModel"),
			"vehicleNumber":         get("vehicleNumber"),
			"driverLicenseNumber":   get("driverLicenseNumber"),
			"pfpUrl":                saveFile("pfp"),
			"driverLicensePhotoUrl": saveFile("driverLicensePhoto"),
			"rcPhotoUrl":            saveFile("rcPhoto"),
		}
	}

	var result *services.LoginResult
	var err error

	// If driver details are present, keep the user type as Driver.
	if dl, ok := loginData["driverLicenseNumber"].(string); ok && dl != "" {
		result, err = services.DriverLoginService(loginData)
	} else if vn, ok := loginData["vehicleNumber"].(string); ok && vn != "" {
		result, err = services.DriverLoginService(loginData)
	} else {
		result, err = services.CustomerLoginService(loginData)
	}
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

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	updatedUser, err := services.UpdateDriverProfile(user.ID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updatedUser)
}

func UpdateDriverLocation(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	if user.UserType != "Driver" {
		http.Error(w, "Forbidden: not a driver", http.StatusForbidden)
		return
	}

	var req struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the driver ID associated with the user ID
	var driver models.Driver
	if err := services.DB.Where("user_id = ?", user.ID).First(&driver).Error; err != nil {
		http.Error(w, "Driver profile not found", http.StatusNotFound)
		return
	}

	// Call the existing service to update location. For HTTP requests, we always persist.
	services.UpdateDriverLocation(driver.ID, req.Lat, req.Lng, true)

	w.WriteHeader(http.StatusOK)
}
