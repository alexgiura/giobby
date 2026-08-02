package models

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    string `json:"code"`              // Error code for programmatic handling
	Message string `json:"message"`           // Human-readable error message
	Details string `json:"details,omitempty"` // Additional details (only in dev mode)
}

