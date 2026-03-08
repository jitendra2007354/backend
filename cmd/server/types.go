package main

// APIResponse is a standard response structure for API endpoints
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationQuery holds common pagination parameters
type PaginationQuery struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
