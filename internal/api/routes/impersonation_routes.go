package routes

import (
	"net/http"
	"superaib/internal/api/handlers"

	"github.com/gorilla/mux"
)

func ImpersonationRoutes(router *mux.Router, handler *handlers.ImpersonationHandler, apiKeyAuth func(http.Handler) http.Handler) {

	// -----------------------------------------------------------
	// 1. USER SPECIFIC ROUTES (Create, Revoke, Extend)
	// URL: /projects/{reference_id}/auth-users/{user_id}/impersonate
	// -----------------------------------------------------------
	userRoute := router.PathPrefix("/projects/{reference_id}/auth-users/{user_id}/impersonate").Subrouter()
	userRoute.Use(apiKeyAuth) // Waa qasab API Key

	// Create Token (Isha)
	userRoute.HandleFunc("", handler.CreateToken).Methods("POST")
	// Get Active Tokens
	userRoute.HandleFunc("/active", handler.GetActiveTokens).Methods("GET")
	// Revoke Specific Token
	userRoute.HandleFunc("/{token_id}/revoke", handler.RevokeToken).Methods("POST")
	// Extend Specific Token
	userRoute.HandleFunc("/{token_id}/extend", handler.ExtendToken).Methods("POST")

	// -----------------------------------------------------------
	// 2. ✅ PROJECT WIDE ROUTES (Validation) - KANI WAA KII MAQNAA
	// URL: /projects/{reference_id}/auth-users/impersonate/validate
	// -----------------------------------------------------------
	validateRoute := router.PathPrefix("/projects/{reference_id}/auth-users/impersonate/validate").Subrouter()
	// Demo App-ka ayaa wacaya, markaa API Key looma baahna haddii Token-ku isagu is-xaqiijinayo,
	// laakiin ammaanka aawadiis waad u deyn kartaa ama waad ka qaadi kartaa.
	// Halkan waxaan u deynaynaa la'aan API Key middleware haddii Demo App-ku uusan dirin x-api-key qaybtan.
	// Laakiin SDK-gaaga waan ku darnay 'x-api-key', markaa waan u deynaynaa.

	validateRoute.HandleFunc("", handler.ValidateToken).Methods("POST")
}
