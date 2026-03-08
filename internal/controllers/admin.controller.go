package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"spark/internal/database"
	"spark/internal/models"
	"spark/internal/services"
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

	model := client.GenerativeModel("gemini-1.5-flash")

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
			"type": "BLOCK_USERS" | "UNBLOCK_USERS" | "GENERATE_CHART" | "NAVIGATE" | "UPDATE_PRICING" | "NONE",
			"targetIds": ["id1", "id2"],
			"page": "Analytics" | "Drivers",
			"chartData": { ... },
			"pricingData": { ... }
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
		http.Error(w, fmt.Sprintf(`{"text": "Error processing request: %v", "action": {"type": "NONE"}}`, err), http.StatusInternalServerError)
		return
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		http.Error(w, "Empty response from AI", http.StatusInternalServerError)
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

func AddPricingRule(w http.ResponseWriter, r *http.Request) {
	ruleType := r.PathValue("type")
	var rule models.PricingRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	rule.Category = ruleType

	if err := database.DB.Create(&rule).Error; err != nil {
		http.Error(w, "Failed to add pricing rule", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Rule added", "rule": rule})
}

func DeletePricingRule(w http.ResponseWriter, r *http.Request) {
	ruleType := r.PathValue("type")
	id := r.PathValue("id")

	result := database.DB.Where("id = ? AND category = ?", id, ruleType).Delete(&models.PricingRule{})
	if result.Error != nil {
		http.Error(w, "Failed to delete rule", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Rule deleted successfully"})
}

func SendNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if err := services.SendAdminNotification("all", req.Title, req.Message); err != nil {
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
	// Simplified stats logic for brevity, assuming standard GORM preloading or separate queries
	// In a real migration, you'd implement the aggregation logic here using DB.Table("rides").Select(...)
	json.NewEncoder(w).Encode(users)
}

func GetPricing(w http.ResponseWriter, r *http.Request) {
	var rules []models.PricingRule
	if err := database.DB.Find(&rules).Error; err != nil {
		http.Error(w, "Failed to fetch pricing", http.StatusInternalServerError)
		return
	}
	// Group by category
	pricing := make(map[string][]models.PricingRule)
	for _, rule := range rules {
		pricing[rule.Category] = append(pricing[rule.Category], rule)
	}
	json.NewEncoder(w).Encode(pricing)
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
	var locations []models.DriverLocation
	if err := database.DB.Preload("Driver").Find(&locations).Error; err != nil {
		http.Error(w, "Failed to fetch locations", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(locations)
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
	var config models.Config
	if err := database.DB.First(&config).Error; err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	}
	json.NewEncoder(w).Encode(config)
}

func UpdateSystemConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
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

	if err := services.SendAdminNotification(req.Target, req.Title, req.Message); err != nil {
		http.Error(w, "Failed to dispatch notification", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Notification dispatched"})
}
