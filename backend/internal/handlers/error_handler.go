package handlers

import (
	"encoding/json"
	"dnsc_microservice/internal/models"
	"log"
	"net/http"
	"os"
	"strings"
)

// Error codes
const (
	ErrCodeInvalidRequest       = "INVALID_REQUEST"
	ErrCodeValidationFailed      = "VALIDATION_FAILED"
	ErrCodeNotFound             = "NOT_FOUND"
	ErrCodeConflict             = "CONFLICT"
	ErrCodeForeignKeyViolation  = "FOREIGN_KEY_VIOLATION"
	ErrCodeUnauthorized         = "UNAUTHORIZED"
	ErrCodeForbidden            = "FORBIDDEN"
	ErrCodeInternalError        = "INTERNAL_ERROR"
)

// isDevelopment checks if the application is running in development mode
// Checks both ENVIRONMENT and DEBUG_MODE from .env file
func isDevelopment() bool {
	env := os.Getenv("ENVIRONMENT")
	debugMode := os.Getenv("DEBUG_MODE")
	
	// Check ENVIRONMENT variable
	isDevEnv := env == "development" || env == "dev" || env == ""
	
	// Check DEBUG_MODE variable (if set to "true" or "1")
	isDebugMode := debugMode == "true" || debugMode == "1"
	
	// Return true if either condition is met
	return isDevEnv || isDebugMode
}

// respondWithError sends a standardized error response
// Technical details are only included in development mode
// Status code is set in HTTP header, not in response body
func respondWithError(w http.ResponseWriter, statusCode int, code, message, details string) {
	errorResponse := models.ErrorResponse{
		Code:    code,
		Message: message,
	}

	// Only include details in development/debug mode
	// In production, hide technical details for security
	if isDevelopment() && details != "" {
		errorResponse.Details = details
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}

// parseDatabaseError converts database errors to user-friendly messages
func parseDatabaseError(err error) (statusCode int, code, message string) {
	if err == nil {
		return http.StatusInternalServerError, ErrCodeInternalError, "An unexpected error occurred"
	}

	errStr := err.Error()

	// Foreign key violations
	if strings.Contains(errStr, "foreign key constraint") {
		if strings.Contains(errStr, "tenant_id") {
			return http.StatusBadRequest, ErrCodeForeignKeyViolation,
				"Invalid tenant ID. The specified tenant does not exist."
		}
		if strings.Contains(errStr, "id_um") {
			return http.StatusBadRequest, ErrCodeForeignKeyViolation,
				"Invalid unit of measure. The specified unit does not exist."
		}
		if strings.Contains(errStr, "id_vat") {
			return http.StatusBadRequest, ErrCodeForeignKeyViolation,
				"Invalid VAT rate. The specified VAT rate does not exist."
		}
		if strings.Contains(errStr, "id_category") {
			return http.StatusBadRequest, ErrCodeForeignKeyViolation,
				"Invalid category. The specified category does not exist."
		}
		return http.StatusBadRequest, ErrCodeForeignKeyViolation,
			"Invalid reference. One of the referenced resources does not exist."
	}

	// Unique constraint violations
	if strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key") {
		return http.StatusConflict, ErrCodeConflict,
			"A resource with this identifier already exists."
	}

	// Not found errors
	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "no rows") {
		return http.StatusNotFound, ErrCodeNotFound,
			"The requested resource was not found."
	}

	// Default to internal error
	return http.StatusInternalServerError, ErrCodeInternalError,
		"An unexpected error occurred. Please try again later."
}

