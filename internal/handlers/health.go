package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	logger *logrus.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *pgxpool.Pool, redisClient *redis.Client, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redisClient,
		logger: logger,
	}
}

// Health performs a basic health check
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "re9ai-whatsapp-adapter",
		"version":   "1.0.0",
	})
}

// Ready performs a readiness check including database and Redis connectivity
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := "ready"
	statusCode := http.StatusOK
	checks := make(map[string]interface{})

	// Check database connectivity
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			h.logger.WithError(err).Error("Database health check failed")
			checks["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			status = "not ready"
			statusCode = http.StatusServiceUnavailable
		} else {
			checks["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	} else {
		checks["database"] = map[string]interface{}{
			"status": "not configured",
		}
	}

	// Check Redis connectivity
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			h.logger.WithError(err).Error("Redis health check failed")
			checks["redis"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			status = "not ready"
			statusCode = http.StatusServiceUnavailable
		} else {
			checks["redis"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	} else {
		checks["redis"] = map[string]interface{}{
			"status": "not configured",
		}
	}

	c.JSON(statusCode, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"service":   "re9ai-whatsapp-adapter",
		"version":   "1.0.0",
		"checks":    checks,
	})
}

// PrometheusHandler returns a handler for Prometheus metrics
func PrometheusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement Prometheus metrics
		c.String(http.StatusOK, "# Prometheus metrics endpoint\n# TODO: Implement metrics collection\n")
	}
}
