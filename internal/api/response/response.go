package response

import (
	"encoding/json"
	"net/http"
	"superaib/internal/core/logger"
)

// APIResponse structure for consistent API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"` // For detailed errors
}

// JSON sends a consistent JSON response
func JSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Success: statusCode >= 200 && statusCode < 300,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Log.Errorf("Failed to write JSON response: %v", err)
		// Fallback for an error writing the response itself
		http.Error(w, "Failed to write JSON response", http.StatusInternalServerError)
	}
}

// Error sends an error JSON response
func Error(w http.ResponseWriter, statusCode int, message string, details ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Success: false,
		Message: message,
	}

	if len(details) > 0 {
		resp.Error = details[0] // Use the first detail as a more specific error object
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Log.Errorf("Failed to write error JSON response: %v", err)
		http.Error(w, "Failed to write error JSON response", http.StatusInternalServerError)
	}
	logger.Log.WithField("status", statusCode).WithField("message", message).WithField("details", details).Warnf("API Error: %s", message)
}
