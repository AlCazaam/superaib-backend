package main

import (
	"context"
	"fmt"
	"net/http"
	"superaib/internal/api/handlers"
	"superaib/internal/api/middleware"
	"superaib/internal/api/routes"
	"superaib/internal/core/config"
	"superaib/internal/core/logger"
	"superaib/internal/models"
	"superaib/internal/services"
	"superaib/internal/storage/postgres"
	"superaib/internal/storage/repo"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// 1. Initialize
	logger.Init()
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatalf("Failed to load config: %v", err)
	}

	db, err := postgres.NewDB(cfg)
	if err != nil {
		logger.Log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	// 2. Migrate & Seed
	logger.Log.Info("Starting database migration...")
	if err := db.DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.APIKey{},
		&models.ProjectFeature{},
		&models.AuthUser{},
		&models.Collection{},
		&models.Document{},
		&models.RealtimeChannel{},
		&models.RealtimeEvent{},
		&models.StorageFile{},
		&models.Analytics{},
		&models.ProjectUsage{},
		&models.GlobalFeature{},
		&models.GlobalAuthProvider{},
		&models.ProjectAuthConfig{},
		&models.ImpersonationToken{},
		&models.PasswordResetToken{},
		&models.RateLimitPolicy{},
		&models.OtpUserTracker{},
		&models.Plan{},
		&models.Transaction{},
		&models.DeviceToken{},
		&models.Notification{},
		&models.ProjectPushConfig{},
	); err != nil {
		logger.Log.Fatalf("Failed to migrate models: %v", err)
	}
	logger.Log.Info("Database migration completed.")

	models.SeedGlobalFeatures(db.DB)
	logger.Log.Info("✅ Global features seeded successfully.")

	// 3. Repositories
	userRepo := repo.NewGormUserRepository(db.DB)
	projectRepo := repo.NewGormProjectRepository(db.DB)
	apiKeyRepo := repo.NewGormAPIKeyRepository(db.DB)
	authUserRepo := repo.NewAuthUserRepository(db.DB)
	documentRepo := repo.NewDocumentRepository(db.DB)
	realtimeChannelRepo := repo.NewGormRealtimeChannelRepository(db.DB)
	realtimeEventRepo := repo.NewGormRealtimeEventRepository(db.DB)
	storageRepo := repo.NewGormStorageRepository(db.DB)
	analyticsRepo := repo.NewGormAnalyticsRepository(db.DB)
	usageRepo := repo.NewGormProjectUsageRepository(db.DB)
	featureRepo := repo.NewGormProjectFeatureRepo(db.DB)
	globalFeatureRepo := repo.NewGormGlobalFeatureRepo(db.DB)
	authProvRepo := repo.NewGlobalAuthProviderRepo(db.DB)
	projectAuthConfigRepo := repo.NewProjectAuthConfigRepo(db.DB)
	impersonationRepo := repo.NewImpersonationRepository(db.DB)
	passResetRepo := repo.NewPasswordResetRepository(db.DB)
	rateLimitRepo := repo.NewRateLimitPolicyRepository(db.DB)
	otpTrackerRepo := repo.NewOtpTrackerRepository(db.DB)
	planRepo := repo.NewPlanRepository(db.DB)
	transactionRepo := repo.NewTransactionRepository(db.DB)
	noteRepo := repo.NewNotificationRepository(db.DB)
	pushConfigRepo := repo.NewPushConfigRepository(db.DB)

	// 4. Services
	analyticsService := services.NewAnalyticsService(analyticsRepo)
	analyticsTracker := analyticsService.GetTracker()
	usageService := services.NewProjectUsageService(usageRepo)
	featureService := services.NewProjectFeatureService(featureRepo, db.DB)
	globalFeatureService := services.NewGlobalFeatureService(globalFeatureRepo)
	authProvService := services.NewGlobalAuthProviderService(authProvRepo)
	otpTrackerService := services.NewOtpTrackerService(otpTrackerRepo, rateLimitRepo, authUserRepo)
	documentService := services.NewDocumentService(documentRepo, analyticsTracker, usageService)
	projectService := services.NewProjectService(projectRepo, featureService, analyticsService, usageService, db.DB)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo, projectRepo, analyticsTracker, usageService)
	authUserService := services.NewAuthUserService(authUserRepo, projectAuthConfigRepo, analyticsTracker, usageService, db.DB)
	realtimeService := services.NewRealtimeService(realtimeChannelRepo, realtimeEventRepo, analyticsTracker, usageService)
	storageService := services.NewStorageService(storageRepo, featureRepo, analyticsTracker, usageService)
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, cfg)
	projectAuthConfigService := services.NewProjectAuthConfigService(projectAuthConfigRepo)
	impersonationService := services.NewImpersonationService(impersonationRepo, authUserRepo)
	emailService := services.NewEmailService(projectAuthConfigRepo)
	rateLimitService := services.NewRateLimitPolicyService(rateLimitRepo)
	planService := services.NewPlanService(planRepo)
	subscriptionService := services.NewSubscriptionService(db.DB, transactionRepo, projectRepo)
	realtimeHandler := handlers.NewRealtimeHandler(realtimeService)
	noteService := services.NewNotificationService(noteRepo, pushConfigRepo, analyticsTracker, usageService, realtimeHandler)
	// 🚀 KICI SCHEDULER-KA (Background Worker)
	noteService.StartScheduler(context.Background())

	passResetService := services.NewPasswordResetService(
		passResetRepo,
		authUserRepo,
		projectAuthConfigRepo,
		emailService,
		otpTrackerService,
	)

	// 5. Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	projectHandler := handlers.NewProjectHandler(projectService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	documentHandler := handlers.NewDocumentHandler(documentService)
	authUserHandler := handlers.NewAuthUserHandler(authUserService)
	storageHandler := handlers.NewStorageHandler(storageService)
	globalFeatureHandler := handlers.NewGlobalFeatureHandler(globalFeatureService)
	projectFeatureHandler := handlers.NewProjectFeatureHandler(featureService)
	authProvHandler := handlers.NewGlobalAuthProviderHandler(authProvService)
	projectAuthConfigHandler := handlers.NewProjectAuthConfigHandler(projectAuthConfigService)
	impersonationHandler := handlers.NewImpersonationHandler(impersonationService)
	passResetHandler := handlers.NewPasswordResetHandler(passResetService)
	rateLimitHandler := handlers.NewRateLimitPolicyHandler(rateLimitService)
	otpTrackerHandler := handlers.NewOtpTrackerHandler(otpTrackerService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	planHandler := handlers.NewPlanHandler(planService)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)
	noteHandler := handlers.NewNotificationHandler(noteService)
	projectUsageHandler := handlers.NewProjectUsageHandler(usageService, projectService)

	// 6. Middlewares
	authMiddleware := middleware.NewAuthMiddleware(cfg)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(apiKeyService, projectService, usageService)

	// 7. Router Setup
	router := mux.NewRouter()

	// 🚀 A. WEBSOCKET ROUTE (Ugu horreysii router-ka)
	router.HandleFunc("/api/v1/ws/{project_id}", realtimeHandler.HandleWebSocket).Methods("GET")

	// 8. API v1 Subrouter
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// 🚀 B. DEVELOPER AUTH (LOGIN/REGISTER) - Halkan Middleware HA DHIGIN!
	routes.AuthRoutes(apiV1, authHandler)

	// 🚀 C. SDK & API-KEY PROTECTED ROUTES
	sdkRouter := apiV1.PathPrefix("").Subrouter()
	sdkRouter.Use(apiKeyMiddleware.AuthenticateAPIKey)

	routes.AuthUserRoutes(sdkRouter, authUserHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.RealtimeRoutes(sdkRouter, realtimeHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.StorageRoutes(sdkRouter, storageHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.DocumentRoutes(sdkRouter, documentHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.PasswordResetRoutes(sdkRouter, passResetHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.RateLimitRoutes(sdkRouter, rateLimitHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.OtpTrackerRoutes(sdkRouter, otpTrackerHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.NotificationRoutes(sdkRouter, noteHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.ImpersonationRoutes(sdkRouter, impersonationHandler, apiKeyMiddleware.AuthenticateAPIKey)
	routes.ProjectFeatureRoutes(sdkRouter, projectFeatureHandler, apiKeyMiddleware.AuthenticateAPIKey)

	// 🚀 D. DASHBOARD & SETTINGS (JWT JWT PROTECTED)
	dashRouter := apiV1.PathPrefix("").Subrouter()
	dashRouter.Use(authMiddleware.Authenticate)

	routes.UserRoutes(dashRouter, userHandler, authMiddleware.Authenticate)
	routes.ProjectRoutes(dashRouter, projectHandler, authMiddleware.Authenticate)
	routes.APIKeyRoutes(dashRouter, apiKeyHandler, authMiddleware.Authenticate)
	routes.GlobalFeatureRoutes(dashRouter, globalFeatureHandler, authMiddleware)
	routes.GlobalAuthProviderRoutes(dashRouter, authProvHandler, authMiddleware)
	routes.ProjectAuthConfigRoutes(dashRouter, projectAuthConfigHandler, authMiddleware)
	routes.AnalyticsRoutes(dashRouter, analyticsHandler, authMiddleware)
	routes.PlanRoutes(dashRouter, planHandler, authMiddleware)
	routes.SubscriptionRoutes(dashRouter, subscriptionHandler, authMiddleware)
	routes.ProjectUsageRoutes(dashRouter, projectUsageHandler, authMiddleware)

	logger.Log.Info("All routes registered successfully.")

	// 9. CORS SETUP
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Requested-With", "x-api-key", "If-Match", "ETag"},
		ExposedHeaders:   []string{"ETag", "If-Match"},
		AllowCredentials: true,
		Debug:            false,
	})
	handler := c.Handler(router)

	// 10. Server Start
	address := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	logger.Log.Infof("🚀 SuperAIB Backend is live at http://%s", address)

	server := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Log.Fatalf("Server failed to start: %v", err)
	}
}
