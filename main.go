package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/config"
	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/handlers"
	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/middleware"
	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/services"
	"github.com/re9-ai/re9ai-whatsapp-adapter/pkg/database"
	"github.com/re9-ai/re9ai-whatsapp-adapter/pkg/logger"
	"github.com/re9-ai/re9ai-whatsapp-adapter/pkg/redis"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	log.Info("Starting re9.ai WhatsApp Adapter")

	// Initialize database connection
	db, err := database.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis connection
	redisClient, err := redis.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize services
	whatsappService := services.NewWhatsAppService(cfg, log)
	messageService := services.NewMessageService(db, redisClient, log)
	mediaService := services.NewMediaService(cfg, log)
	aiService := services.NewAIService(cfg, log)

	// Initialize handlers
	whatsappHandler := handlers.NewWhatsAppHandler(
		whatsappService,
		messageService,
		mediaService,
		aiService,
		log,
	)
	healthHandler := handlers.NewHealthHandler(db, redisClient, log)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery(log))
	router.Use(middleware.CORS())
	router.Use(middleware.Security())
	router.Use(middleware.RateLimit(redisClient))

	// Health check endpoints
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// WhatsApp webhook endpoints
	whatsappGroup := router.Group("/webhooks/whatsapp")
	{
		whatsappGroup.GET("/verify", whatsappHandler.VerifyWebhook)
		whatsappGroup.POST("/messages", 
			middleware.WhatsAppSignatureVerification(cfg.WhatsAppWebhookSecret),
			whatsappHandler.HandleMessage,
		)
		whatsappGroup.POST("/status", 
			middleware.WhatsAppSignatureVerification(cfg.WhatsAppWebhookSecret),
			whatsappHandler.HandleStatus,
		)
	}

	// API endpoints for internal communication
	apiGroup := router.Group("/api/v1")
	{
		apiGroup.POST("/messages/send", whatsappHandler.SendMessage)
		apiGroup.GET("/messages/:messageId", whatsappHandler.GetMessage)
		apiGroup.POST("/media/upload", whatsappHandler.UploadMedia)
	}

	// Metrics endpoint for Prometheus
	router.GET("/metrics", handlers.PrometheusHandler())

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}