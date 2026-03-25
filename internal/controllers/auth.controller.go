package controllers

import (
	"encoding/json"
	"net/http"
	"spark/internal/middleware"
	"spark/internal/models"
	"spark/internal/services"
	"strings"
)

// writeAuthJsonError is a helper to format error responses as JSON.
func writeAuthJsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	loginData := map[string]interface{}{}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
			writeAuthJsonError(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Handle form-data fallback for file uploads and legacy clients
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeAuthJsonError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Helper to save file and return URL
		saveFile := func(field string) string {
			file, header, err := r.FormFile(field)
			if err != nil {
				// Fallback: If it's not a file, check if it's a direct URL string (e.g., from Google Sign-In)
				val := r.FormValue(field)
				if val != "" {
					return val
				}
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
			"rcNumber":              get("rcNumber"),
			"driverLicenseNumber":   get("driverLicenseNumber"),
			"pfpUrl":                saveFile("pfp"),
			"driverLicensePhotoUrl": saveFile("driverLicensePhoto"),
			"rcPhotoUrl":            saveFile("rcPhoto"),
			"userType":              get("userType"),
			"action":                get("action"),
		}
	}

	var result *services.LoginResult
	var err error

	// Route to Driver login if explicitly requested, or if typical driver fields are present
	userType, _ := loginData["userType"].(string)
	dl, _ := loginData["driverLicenseNumber"].(string)
	vn, _ := loginData["vehicleNumber"].(string)

	if userType == "Driver" || dl != "" || vn != "" {
		result, err = services.DriverLoginService(loginData)
	} else {
		result, err = services.CustomerLoginService(loginData)
	}

	if err != nil {
		status := http.StatusInternalServerError
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "registered") || strings.Contains(errMsg, "required") {
			status = http.StatusUnauthorized
		}
		writeAuthJsonError(w, errMsg, status)
		return
	}

	userMap := map[string]interface{}{}
	userBytes, _ := json.Marshal(result.User)
	json.Unmarshal(userBytes, &userMap)

	if result.User != nil && result.User.UserType == "Driver" {
		if vehicles, err := services.ListDriverVehicles(result.User.ID); err == nil {
			userMap["vehicles"] = vehicles
		} else {
			userMap["vehicles"] = []interface{}{}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   result.Token,
		"user":    userMap,
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
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userMap := map[string]interface{}{}
	userBytes, _ := json.Marshal(user)
	json.Unmarshal(userBytes, &userMap)

	if user.UserType == "Driver" {
		if vehicles, err := services.ListDriverVehicles(user.ID); err == nil {
			userMap["vehicles"] = vehicles
		} else {
			userMap["vehicles"] = []interface{}{}
		}
	}

	json.NewEncoder(w).Encode(userMap)
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
