package main

import (
	"context"
	"fmt"
	"home-server-dashboard/backend/internal/config"
	"home-server-dashboard/backend/internal/database"
	"home-server-dashboard/backend/internal/handler"
	"home-server-dashboard/backend/internal/middleware"
	"home-server-dashboard/backend/internal/repository"
	"home-server-dashboard/backend/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Configuration error: ", err)
	}

	log.Printf("Starting in %s environment", cfg.Environment)

	// Database
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL successfully")

	// Repositories
	authRepo := repository.NewAuthRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	serverRepo := repository.NewServerRepository(db)
	metricsRepo := repository.NewMetricsRepository(db)
	agentTokenRepo := repository.NewAgentTokenRepository(db)

	// Services

	tokenService := service.NewTokenService(
		cfg.JWTSecret,
		cfg.AccessTokenMinutes,
		cfg.RefreshTokenDays,
	)
	authService := service.NewAuthService(authRepo, tokenService)
	var passwordResetMailer service.PasswordResetMailer
	if cfg.SMTPHost != "" {
		passwordResetMailer = service.NewSMTPPasswordResetMailer(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.SMTPFrom,
		)
	} else if cfg.Environment == config.Production {
		passwordResetMailer = service.UnavailablePasswordResetMailer{}
	} else {
		passwordResetMailer = service.NewConsolePasswordResetMailer(log.Default())
		log.Println("SMTP is not configured; password reset codes will appear in this development console")
	}
	passwordResetService := service.NewPasswordResetService(
		authRepo,
		passwordResetRepo,
		passwordResetMailer,
		cfg.JWTSecret,
	)
	serverService := service.NewServerService(serverRepo)
	metricsService := service.NewMetricsService(serverRepo, metricsRepo)
	dashboardService := service.NewDashboardService(
		serverRepo,
		metricsRepo,
	)

	agentTokenService := service.NewAgentTokenService(agentTokenRepo)

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	agentMiddleware := middleware.NewAgentMiddleware(agentTokenService)

	// Handlers

	authHandler := handler.NewAuthHandler(authService)
	passwordResetHandler := handler.NewPasswordResetHandler(passwordResetService)
	serverHandler := handler.NewServerHandler(serverService)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	dashboardHandler := handler.NewDashboardHandler(
		dashboardService,
	)
	agentTokenHandler := handler.NewAgentTokenHandler(
		serverService,
		agentTokenService,
	)

	userHandler := handler.NewUserHandler(authService)

	// Routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)

	mux.Handle(
		"POST /api/v1/servers/{server_id}/agent-token",
		authMiddleware.Authenticate(
			http.HandlerFunc(agentTokenHandler.Generate),
		),
	)

	mux.Handle(
		"DELETE /api/v1/servers/{server_id}/agent-token",
		authMiddleware.Authenticate(
			http.HandlerFunc(agentTokenHandler.Revoke),
		),
	)

	mux.Handle("GET /api/v1/servers",
		authMiddleware.Authenticate(
			http.HandlerFunc(serverHandler.GetServers),
		),
	)
	mux.Handle("POST /api/v1/servers",
		authMiddleware.Authenticate(
			http.HandlerFunc(serverHandler.CreateServer)))
	mux.Handle("DELETE /api/v1/servers/{server_id}",
		authMiddleware.Authenticate(
			http.HandlerFunc(serverHandler.DeleteServer)))

	mux.Handle("POST /api/v1/metrics",
		agentMiddleware.Authenticate(
			http.HandlerFunc(metricsHandler.Create)))

	mux.Handle("GET /api/v1/servers/{server_id}/metrics",
		authMiddleware.Authenticate(
			http.HandlerFunc(metricsHandler.GetByServerID),
		),
	)

	mux.Handle("GET /api/v1/servers/{server_id}", authMiddleware.Authenticate(
		http.HandlerFunc(serverHandler.GetByID),
	),
	)

	mux.Handle("GET /api/v1/servers/{server_id}/summary",
		authMiddleware.Authenticate(
			http.HandlerFunc(dashboardHandler.GetSummary),
		),
	)

	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc(
		"POST /api/v1/auth/password-reset/request",
		passwordResetHandler.RequestCode,
	)
	mux.HandleFunc(
		"POST /api/v1/auth/password-reset/confirm",
		passwordResetHandler.Confirm,
	)

	mux.Handle(
		"GET /api/v1/auth/me",
		authMiddleware.Authenticate(
			http.HandlerFunc(authHandler.Me),
		),
	)
	mux.Handle(
		"PATCH /api/v1/users/me",
		authMiddleware.Authenticate(
			http.HandlerFunc(userHandler.UpdateProfile),
		),
	)

	// HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Home Server Dashboard API starting on :%d", cfg.Port)
	retentionService := service.NewRetentionService(
		repository.NewRetentionRepository(db),
		log.Default(),
	)
	go retentionService.Run(context.Background())

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
