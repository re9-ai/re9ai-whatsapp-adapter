package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresConnection creates a new PostgreSQL connection pool
func NewPostgresConnection(databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// CreateTables creates the necessary database tables for the WhatsApp adapter
func CreateTables(ctx context.Context, db *pgxpool.Pool) error {
	// Create whatsapp_messages table
	createMessagesTable := `
	CREATE TABLE IF NOT EXISTS whatsapp_messages (
		id UUID PRIMARY KEY,
		twilio_sid VARCHAR(255) UNIQUE NOT NULL,
		from_number VARCHAR(50) NOT NULL,
		to_number VARCHAR(50) NOT NULL,
		direction VARCHAR(20) NOT NULL CHECK (direction IN ('inbound', 'outbound')),
		message_type VARCHAR(20) NOT NULL CHECK (message_type IN ('text', 'image', 'document', 'audio', 'video', 'location', 'contact')),
		status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'sent', 'delivered', 'read', 'failed')),
		content TEXT,
		media_url TEXT,
		media_type VARCHAR(100),
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		user_id UUID,
		session_id UUID,
		error_code VARCHAR(50),
		error_message TEXT
	);`

	if _, err := db.Exec(ctx, createMessagesTable); err != nil {
		return fmt.Errorf("failed to create whatsapp_messages table: %w", err)
	}

	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS whatsapp_users (
		id UUID PRIMARY KEY,
		phone_number VARCHAR(50) UNIQUE NOT NULL,
		whatsapp_id VARCHAR(100) UNIQUE,
		profile_name VARCHAR(255),
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	if _, err := db.Exec(ctx, createUsersTable); err != nil {
		return fmt.Errorf("failed to create whatsapp_users table: %w", err)
	}

	// Create chat_sessions table
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS chat_sessions (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL REFERENCES whatsapp_users(id),
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		context JSONB,
		started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		ended_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	if _, err := db.Exec(ctx, createSessionsTable); err != nil {
		return fmt.Errorf("failed to create chat_sessions table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_messages_from_number ON whatsapp_messages(from_number);",
		"CREATE INDEX IF NOT EXISTS idx_messages_to_number ON whatsapp_messages(to_number);",
		"CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON whatsapp_messages(timestamp);",
		"CREATE INDEX IF NOT EXISTS idx_messages_status ON whatsapp_messages(status);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON chat_sessions(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_status ON chat_sessions(status);",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
