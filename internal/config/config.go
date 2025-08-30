package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the WhatsApp adapter service
type Config struct {
	// Server configuration
	Port        string
	Environment string
	LogLevel    string

	// Database configuration
	DatabaseURL string
	RedisURL    string

	// Twilio configuration
	TwilioAccountSID       string
	TwilioAuthToken        string
	TwilioWhatsAppFrom     string // e.g., "whatsapp:+14155238886"
	
	// WhatsApp webhook configuration
	WhatsAppWebhookSecret  string
	WhatsAppVerifyToken    string

	// AWS configuration for media handling
	AWSRegion           string
	AWSAccessKeyID      string
	AWSSecretAccessKey  string
	S3BucketName        string

	// External service URLs
	ChatOrchestratorURL string
	AIProcessingURL     string

	// Rate limiting
	RateLimitPerMinute int
	RateLimitBurst     int

	// Security
	JWTSecret string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server configuration
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		// Database configuration
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),

		// Twilio configuration
		TwilioAccountSID:       getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:        getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioWhatsAppFrom:     getEnv("TWILIO_WHATSAPP_FROM", "whatsapp:+14155238886"),

		// WhatsApp webhook configuration
		WhatsAppWebhookSecret:  getEnv("WHATSAPP_WEBHOOK_SECRET", ""),
		WhatsAppVerifyToken:    getEnv("WHATSAPP_VERIFY_TOKEN", ""),

		// AWS configuration
		AWSRegion:           getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:      getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:  getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3BucketName:        getEnv("S3_BUCKET_NAME", ""),

		// External service URLs
		ChatOrchestratorURL: getEnv("CHAT_ORCHESTRATOR_URL", "http://localhost:8081"),
		AIProcessingURL:     getEnv("AI_PROCESSING_URL", "http://localhost:8082"),

		// Rate limiting
		RateLimitPerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
		RateLimitBurst:     getEnvAsInt("RATE_LIMIT_BURST", 10),

		// Security
		JWTSecret: getEnv("JWT_SECRET", ""),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// Validate checks if all required configuration values are set
func (c *Config) Validate() error {
	required := map[string]string{
		"TWILIO_ACCOUNT_SID":      c.TwilioAccountSID,
		"TWILIO_AUTH_TOKEN":       c.TwilioAuthToken,
		"WHATSAPP_WEBHOOK_SECRET": c.WhatsAppWebhookSecret,
		"WHATSAPP_VERIFY_TOKEN":   c.WhatsAppVerifyToken,
		"DATABASE_URL":            c.DatabaseURL,
		"JWT_SECRET":              c.JWTSecret,
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	return nil
}