package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"spark/internal/database"
	"spark/internal/models"
	"spark/internal/services"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// --- AI ASSISTANT --- //

type AIRequest struct {
	Prompt     string                   `json:"prompt"`
	History    []map[string]interface{} `json:"history"`
	UserData   []map[string]interface{} `json:"userData"`
	FileBase64 string                   `json:"fileBase64"`
	MimeType   string                   `json:"mimeType"`
}

func GetAIAssistantResponse(w http.ResponseWriter, r *http.Request) {
	var reqBody AIRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		http.Error(w, `{"text": "AI Assistant is disabled. Configure GEMINI_API_KEY.", "action": {"type": "NONE"}}`, http.StatusInternalServerError)
		return
	}

	if reqBody.Prompt == "" {
		http.Error(w, `{"error": "Prompt is required."}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		http.Error(w, "Failed to create AI client", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.0-pro")

	// Simplify user data for context
	simplifiedData := make([]map[string]interface{}, 0)
	for i, u := range reqBody.UserData {
		if i >= 100 {
			break
		}
		simplifiedData = append(simplifiedData, map[string]interface{}{
			"id":         u["id"],
			"role":       u["role"],
			"city":       u["city"],
			"state":      u["state"],
			"earnings":   u["earnings"],
			"revenue":    u["revenue"],
			"bookings":   u["bookings"],
			"penalty":    u["penalty"],
			"status":     u["status"],
			"signupDate": u["signupDate"],
		})
	}
	userDataJSON, _ := json.Marshal(simplifiedData)

	systemInstructionText := fmt.Sprintf(`
		You are the intelligent AI Admin Assistant for 'Spark'.
		You must return a JSON object. Do NOT wrap it in markdown blocks. Just raw JSON.
		Structure:
		{
		"text": "Your conversational response here.",
		"action": {
			"type": "BLOCK_USERS" | "UNBLOCK_USERS" | "GENERATE_CHART" | "NAVIGATE" | "UPDATE_CONFIG" | "NONE",
			"targetIds": ["id1", "id2"],
			"page": "Analytics" | "Drivers",
			"chartData": { ... },
			"configData": { "platformFee": 50, "gstPercentage": 18, "maxBidPerKm": 20, "cancellationFee": 0 }
		}
		}
		Current Date: %s
		Current App Data: %s
	`, time.Now().Format("2006-01-02"), string(userDataJSON))

	// Construct chat session
	cs := model.StartChat()
	cs.History = []*genai.Content{
		{
			Role:  "user",
			Parts: []genai.Part{genai.Text(systemInstructionText)},
		},
	}

	// Add previous history
	for _, msg := range reqBody.History {
		role := "user"
		if r, ok := msg["sender"].(string); ok && r != "user" {
			role = "model"
		}
		text, _ := msg["text"].(string)
		cs.History = append(cs.History, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(text)},
		})
	}

	// Current prompt
	parts := []genai.Part{genai.Text(reqBody.Prompt)}
	if reqBody.FileBase64 != "" && reqBody.MimeType != "" {
		parts = append(parts, genai.Blob{
			MIMEType: reqBody.MimeType,
			Data:     []byte(reqBody.FileBase64), // Assumes decoded or client handles decoding if needed, usually SDK handles raw bytes
		})
	}

	resp, err := cs.SendMessage(ctx, parts...)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"text": "Error processing request: %v", "action": {"type": "NONE"}}`, err)))
		return
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"text": "Empty response from AI", "action": {"type": "NONE"}}`))
		return
	}

	// Assuming text response
	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		cleanText := strings.TrimSpace(string(txt))
		cleanText = strings.TrimPrefix(cleanText, "```json")
		cleanText = strings.TrimPrefix(cleanText, "```")
		cleanText = strings.TrimSuffix(cleanText, "```")
		cleanText = strings.TrimSpace(cleanText)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cleanText))
	} else {
		http.Error(w, "Unexpected response format", http.StatusInternalServerError)
	}
}

func SendNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target  string `json:"target"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	target := req.Target
	lowerTarget := strings.ToLower(target)
	if lowerTarget == "" || lowerTarget == "all" {
		target = "all"
	} else if lowerTarget == "customer" {
		target = "Customer"
	} else if lowerTarget == "driver" {
		target = "Driver"
	} else if lowerTarget == "campowner" {
		target = "CampOwner"
	}

	if err := services.SendAdminNotification(target, req.Title, req.Message); err != nil {
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification sent successfully"})
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := database.DB.Omit("PFP").Find(&users).Error; err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	var rawUsers []map[string]interface{}
	userBytes, _ := json.Marshal(users)
	json.Unmarshal(userBytes, &rawUsers)

	for i, u := range rawUsers {
		// Map userType to role
		var roleStr string
		if uType, ok := u["userType"].(string); ok {
			roleStr = uType
		} else if uType, ok := u["UserType"].(string); ok {
			roleStr = uType
		}

		if strings.EqualFold(roleStr, "driver") {
			rawUsers[i]["role"] = "Driver"
		} else if strings.EqualFold(roleStr, "customer") {
			rawUsers[i]["role"] = "Customer"
		} else if strings.EqualFold(roleStr, "campowner") {
			rawUsers[i]["role"] = "CampOwner"
		} else {
			rawUsers[i]["role"] = roleStr
		}

		// Map firstName and lastName to name
		fName, _ := u["firstName"].(string)
		lName, _ := u["lastName"].(string)
		if fName == "" && u["FirstName"] != nil {
			fName, _ = u["FirstName"].(string)
			lName, _ = u["LastName"].(string)
		}
		name := fName
		if lName != "" {
			name += " " + lName
		}
		rawUsers[i]["name"] = name

		// Map phoneNumber to mobile
		if phone, ok := u["phoneNumber"]; ok {
			rawUsers[i]["mobile"] = phone
		} else if phone, ok := u["PhoneNumber"]; ok {
			rawUsers[i]["mobile"] = phone
		}

		// Set default status if missing
		isBlocked := false
		if b, ok := u["isBlocked"].(bool); ok {
			isBlocked = b
		} else if b, ok := u["IsBlocked"].(bool); ok {
			isBlocked = b
		}
		if isBlocked {
			rawUsers[i]["status"] = "Blocked"
		} else {
			rawUsers[i]["status"] = "Active"
		}
	}

	json.NewEncoder(w).Encode(rawUsers)
}

func GetTickets(w http.ResponseWriter, r *http.Request) {
	var tickets []models.SupportTicket
	if err := database.DB.Preload("User").Order("created_at DESC").Find(&tickets).Error; err != nil {
		http.Error(w, "Failed to fetch tickets", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tickets)
}

func GetDriverLocations(w http.ResponseWriter, r *http.Request) {
	type DriverLoc struct {
		UserID    uint    `json:"id"`
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lng"`
	}

	var results []DriverLoc
	err := database.DB.Table("driver_locations").
		Select("drivers.user_id as user_id, driver_locations.latitude as latitude, driver_locations.longitude as longitude").
		Joins("left join drivers on drivers.id = driver_locations.driver_id").
		Scan(&results).Error

	if err != nil || len(results) == 0 {
		var locations []models.DriverLocation
		database.DB.Preload("Driver").Find(&locations)

		var fallbackResults []map[string]interface{}
		for _, loc := range locations {
			locBytes, _ := json.Marshal(loc)
			var locMap map[string]interface{}
			json.Unmarshal(locBytes, &locMap)

			lat := locMap["lat"]
			if lat == nil {
				lat = locMap["latitude"]
			}
			if lat == nil {
				lat = locMap["Latitude"]
			}

			lng := locMap["lng"]
			if lng == nil {
				lng = locMap["longitude"]
			}
			if lng == nil {
				lng = locMap["Longitude"]
			}

			var userId interface{}
			if driver, ok := locMap["Driver"].(map[string]interface{}); ok && driver != nil {
				userId = driver["userId"]
				if userId == nil {
					userId = driver["UserID"]
				}
				if userId == nil {
					userId = driver["user_id"]
				}
			}
			if userId == nil {
				userId = locMap["userId"]
			}
			if userId == nil {
				userId = locMap["UserID"]
			}
			if userId == nil {
				userId = locMap["user_id"]
			}
			if userId == nil {
				userId = locMap["driverId"]
			}
			if userId == nil {
				userId = locMap["DriverID"]
			}
			if userId == nil {
				userId = locMap["id"]
			}

			if lat != nil && lng != nil && userId != nil {
				fallbackResults = append(fallbackResults, map[string]interface{}{
					"id":  fmt.Sprint(userId),
					"lat": lat,
					"lng": lng,
				})
			}
		}
		if fallbackResults == nil {
			fallbackResults = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(fallbackResults)
		return
	}

	json.NewEncoder(w).Encode(results)
}

func GetNotificationHistory(w http.ResponseWriter, r *http.Request) {
	var history []models.Notification
	if err := database.DB.Order("created_at DESC").Find(&history).Error; err != nil {
		http.Error(w, "Failed to fetch history", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(history)
}

func GetSystemConfig(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	if err := database.DB.Model(&models.Config{}).First(&config).Error; err != nil {
		// If no config is in the DB, return a default-filled object
		// to prevent frontends from crashing or showing all zeros.
		json.NewEncoder(w).Encode(map[string]interface{}{
			"platform_fee":          50.0,
			"gst_percentage":        18.0,
			"base_fare":             50.0,
			"base_fare_bike":        25.0,
			"base_fare_auto":        35.0,
			"base_fare_car":         50.0,
			"base_fare_suv":         65.0,
			"base_fare_luxury":      100.0,
			"max_bid_per_km":        30.0,
			"min_bid_per_km":        8.0,
			"min_bid_per_km_bike":   4.8,
			"min_bid_per_km_auto":   6.4,
			"min_bid_per_km_car":    8.0,
			"min_bid_per_km_suv":    9.6,
			"min_bid_per_km_luxury": 14.4,
			"cancellation_fee":      0.0,
			"driver_search_radius":  5.0,
			"ride_accept_time":      60.0,
			"wallet_min_balance":    100.0,
		})
		return
	}

	// Ensure raw byte arrays from DB drivers are converted to readable strings
	for k, v := range config {
		if b, ok := v.([]byte); ok {
			config[k] = string(b)
		}
	}

	json.NewEncoder(w).Encode(config)
}

func UpdateSystemConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	// Convert camelCase keys to snake_case for database columns
	snakeUpdates := make(map[string]interface{})
	for k, v := range updates {
		var snake strings.Builder
		for i, c := range k {
			if i > 0 && c >= 'A' && c <= 'Z' {
				snake.WriteRune('_')
			}
			snake.WriteRune(c)
		}
		snakeUpdates[strings.ToLower(snake.String())] = v
	}

	var config models.Config
	if err := database.DB.First(&config).Error; err != nil {
		database.DB.Create(&config) // Create an initial row if the table is currently empty
	}
	database.DB.Model(&config).Updates(snakeUpdates) // Safely updates specific fields matching snake_case db columns

	if err := services.SetConfig(updates); err != nil {
		http.Error(w, "Failed to update config", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Configuration updated successfully."})
}

func CreateRealtimeNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target  string `json:"target"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	target := req.Target
	lowerTarget := strings.ToLower(target)
	if lowerTarget == "" || lowerTarget == "all" {
		target = "all"
	} else if lowerTarget == "customer" {
		target = "Customer"
	} else if lowerTarget == "driver" {
		target = "Driver"
	} else if lowerTarget == "campowner" {
		target = "CampOwner"
	}

	if err := services.SendAdminNotification(target, req.Title, req.Message); err != nil {
		http.Error(w, "Failed to dispatch notification", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Notification dispatched"})
}
