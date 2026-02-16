package utils

import (
	"errors"
	"net/http"
)

// GetOwnerIDFromContext safely extracts the ownerID from request context
func GetOwnerIDFromContext(r *http.Request) (string, error) {
	v := r.Context().Value("ownerID")
	if id, ok := v.(string); ok && id != "" {
		return id, nil
	}
	return "", errors.New("unauthorized: missing or invalid ownerID in context")
}
