package models

// Response represents standard API success response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents standard API error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
