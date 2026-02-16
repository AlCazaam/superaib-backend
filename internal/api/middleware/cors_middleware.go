package middleware

import (
	"context"
	"net/http"
	"strings"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type contextKey string

const ProjectIDKey contextKey = "projectID"

type APIKeyMiddleware struct {
	apiKeyService  services.APIKeyService
	projectService services.ProjectService
	usageService   services.ProjectUsageService
}

func NewAPIKeyMiddleware(ak services.APIKeyService, ps services.ProjectService, us services.ProjectUsageService) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		apiKeyService:  ak,
		projectService: ps,
		usageService:   us,
	}
}

// AuthenticateAPIKey: Kani waa ilaalada (Gatekeeper) dhamaan SDK Services-ka
func (m *APIKeyMiddleware) AuthenticateAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 🚀 1. WEBSOCKET BYPASS (VERY IMPORTANT)
		// Haddii codsigu yahay WebSocket, Middleware-ku ha dhaafo.
		// Sababta: WebSocket Handshake kuma soo qaado Header-ka "x-api-key".
		// Realtime Handler-ka ayaa isagu si toos ah URL-ka uga hubin doona API Key-ga.
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Hel Project ID (URL Vars, Reference ID ama Query Params)
		vars := mux.Vars(r)
		idOrRefFromURL := vars["project_id"]
		if idOrRefFromURL == "" {
			idOrRefFromURL = vars["reference_id"]
		}
		if idOrRefFromURL == "" {
			idOrRefFromURL = r.URL.Query().Get("project_id")
		}

		if idOrRefFromURL == "" || idOrRefFromURL == "null" {
			response.Error(w, http.StatusBadRequest, "Invalid Project ID", "Project ID is required")
			return
		}

		// 3. Hel API Key (Header "x-api-key" ama Query "api_key")
		apiKeyString := r.Header.Get("x-api-key")
		if apiKeyString == "" {
			apiKeyString = r.URL.Query().Get("api_key")
		}

		if apiKeyString == "" {
			response.Error(w, http.StatusUnauthorized, "Missing API Key", "API key is required for this request")
			return
		}

		// 4. Xaqiiji API Key-ga Database-ka
		apiKey, err := m.apiKeyService.GetAPIKeyByKeyValue(r.Context(), apiKeyString)
		if err != nil || apiKey.Revoked {
			response.Error(w, http.StatusUnauthorized, "Invalid API Key", "The provided API key is invalid or revoked")
			return
		}

		// 5. Hel Project-ka oo xaqiiji jiritaankiisa
		project, err := m.projectService.GetProjectByRefOrID(r.Context(), idOrRefFromURL)
		if err != nil {
			response.Error(w, http.StatusNotFound, "Project not found", "The specified project does not exist")
			return
		}

		// Security Check: Hubi in API Key-gu uu leeyahay Project-kan
		if strings.ToLower(apiKey.ProjectID) != strings.ToLower(project.ID) {
			response.Error(w, http.StatusForbidden, "Security Violation", "API Key mismatch for this project")
			return
		}

		// 6. SMART QUOTA & LIMIT ENFORCEMENT
		usage, err := m.usageService.GetUsage(r.Context(), project.ID)
		if err == nil {
			path := strings.ToLower(r.URL.Path)

			// A. Hubi limit-ka Auth-ka
			if strings.Contains(path, "/auth") && usage.LimitAuthUsers != -1 && usage.AuthUsersCount >= usage.LimitAuthUsers {
				response.Error(w, http.StatusForbidden, "Auth limit reached", "limit_reached_auth")
				return
			}

			// B. Hubi limit-ka Database-ka
			if (strings.Contains(path, "/documents") || strings.Contains(path, "/db")) && usage.LimitDocuments != -1 && usage.DocumentsCount >= usage.LimitDocuments {
				response.Error(w, http.StatusForbidden, "Database limit reached", "limit_reached_db")
				return
			}

			// C. Hubi limit-ka Storage-ka
			if strings.Contains(path, "/storage") && usage.LimitStorageMB != -1 && usage.StorageUsedMB >= usage.LimitStorageMB {
				response.Error(w, http.StatusForbidden, "Storage limit reached", "limit_reached_storage")
				return
			}

			// D. API Calls Quota
			if usage.LimitApiCalls != -1 && usage.ApiCalls >= usage.LimitApiCalls {
				response.Error(w, http.StatusForbidden, "API quota exceeded", "limit_reached_api")
				return
			}
		}

		// 7. UPDATE USAGE (In background si uusan request-ka u gaabin)
		go m.apiKeyService.UpdateAPIKeyUsage(context.Background(), apiKeyString)
		go m.usageService.UpdateUsage(context.Background(), project.ID, "api_calls", 1)

		// 8. Ku dar Project ID Context-ga si Handler-ka dambe u isticmaalo
		ctx := context.WithValue(r.Context(), ProjectIDKey, project.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORSMiddleware: Oggolaanshaha xiriirka dhinaca Front-end (Web/Mobile)
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, x-api-key, If-Match, ETag")
		w.Header().Set("Access-Control-Expose-Headers", "ETag, Content-Type, If-Match")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
