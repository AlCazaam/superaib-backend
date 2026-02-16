package routes

import (
	"superaib/internal/api/handlers"
	"superaib/internal/api/middleware"

	"github.com/gorilla/mux"
)

func PlanRoutes(router *mux.Router, h *handlers.PlanHandler, auth *middleware.AuthMiddleware) {
	adminRouter := router.PathPrefix("/admin/plans").Subrouter()
	adminRouter.Use(auth.Authenticate)

	adminRouter.HandleFunc("", h.CreatePlan).Methods("POST")
	adminRouter.HandleFunc("", h.GetAllPlans).Methods("GET")
	adminRouter.HandleFunc("/{id}", h.UpdatePlan).Methods("PUT")
	adminRouter.HandleFunc("/{id}", h.DeletePlan).Methods("DELETE")
}
