package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// NewRedisClient creates a new Redis client connection
func NewRedisClient(redisURL string) (*redis.Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("Redis URL is required")
	}

	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Create Redis client
	client := redis.NewClient(opt)

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// HealthCheck checks if Redis is accessible
func HealthCheck(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return fmt.Errorf("Redis client is nil")
	}

	return client.Ping(ctx).Err()
}
