package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"spark/internal/services"
)

// writeJsonError is a helper to format error responses as JSON.
func writeJsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// The frontend sends multipart/form-data, so we need to parse it.
	// 10 << 20 sets a 10MB maximum for the entire request body.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJsonError(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Helper to get form value, returns "" if not present, which prevents nil-interface panics.
	get := func(key string) string { return r.FormValue(key) }

	// Explicitly construct the data map for the service. This ensures all expected keys are present,
	// even if the frontend omits empty fields from the FormData.
	registrationData := map[string]interface{}{
		"phoneNumber":         get("phoneNumber"),
		"firstName":           get("firstName"),
		"lastName":            get("lastName"),
		"email":               get("email"),
		"city":                get("city"),
		"state":               get("state"),
		"vehicleType":         get("vehicleType"),
		"vehicleModel":        get("vehicleModel"),
		"vehicleNumber":       get("vehicleNumber"),
		"driverLicenseNumber": get("driverLicenseNumber"),
		"rcNumber":            get("rcNumber"),
		// File URLs will be added below
	}

	// The frontend sends files with field names like 'driverLicensePhoto'.
	// The service layer expects the resulting URL to be in a field named 'driverLicensePhotoUrl'.
	// This map defines that translation.
	fileFieldMapping := map[string]string{
		"pfp":                "pfpUrl",
		"driverLicensePhoto": "driverLicensePhotoUrl",
		"rcPhoto":            "rcPhotoUrl",
	}

	for frontendField, backendField := range fileFieldMapping {
		file, header, err := r.FormFile(frontendField)
		if err == nil { // A file for this field was found
			defer file.Close()
			// Ensure the upload directory exists before saving the file
			uploadDir := "./public/uploads/users"
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
				writeJsonError(w, "Failed to create upload directory: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// The UploadFile service saves the file and returns its public URL.
			url, err := services.UploadFile(file, header.Filename, "users")
			if err != nil {
				writeJsonError(w, "Failed to upload file for field "+frontendField+": "+err.Error(), http.StatusInternalServerError)
				return
			}
			registrationData[backendField] = url
		} else {
			// If the file is not present, ensure the key exists with an empty string
			// to prevent panics in the service laye
			registrationData[backendField] = ""
		}
	}

	// Log the complete data map for debugging before it's passed to the service.
	fmt.Printf("DEBUG: Final Registration Map: %+v\n", registrationData)

	// Select correct service based on driver data flag
	var result *services.LoginResult
	var err error
	if dl, ok := registrationData["driverLicenseNumber"].(string); ok && dl != "" {
		result, err = services.DriverLoginService(registrationData)
	} else {
		result, err = services.CustomerLoginService(registrationData)
	}
	if err != nil {
		writeJsonError(w, "Registration failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure driver has vehicles included in response so frontend can display immediately.
	userMap := map[string]interface{}{}
	userBytes, _ := json.Marshal(result.User)
	json.Unmarshal(userBytes, &userMap)
	if result.User != nil && result.User.UserType == "Driver" {
		if vehicles, err := services.ListDriverVehicles(result.User.ID); err == nil {
			userMap["vehicles"] = vehicles
		} else {
			userMap["vehicles"] = []interface{}{}
		}
	} else {
		userMap["vehicles"] = []interface{}{}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": result.Token,
		"user":  userMap,
	})
}
