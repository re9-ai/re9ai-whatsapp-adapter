package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/models"
)

// MessageService handles message storage and retrieval operations
type MessageService struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	logger *logrus.Logger
}

// NewMessageService creates a new message service instance
func NewMessageService(db *pgxpool.Pool, redisClient *redis.Client, logger *logrus.Logger) *MessageService {
	return &MessageService{
		db:     db,
		redis:  redisClient,
		logger: logger,
	}
}

// StoreMessage stores a WhatsApp message in the database
func (m *MessageService) StoreMessage(ctx context.Context, message *models.WhatsAppMessage) error {
	m.logger.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"twilio_sid":   message.TwilioSID,
		"direction":    message.Direction,
		"message_type": message.Type,
	}).Info("Storing WhatsApp message")

	query := `
		INSERT INTO whatsapp_messages (
			id, twilio_sid, from_number, to_number, direction, message_type, 
			status, content, media_url, media_type, timestamp, created_at, updated_at,
			user_id, session_id, error_code, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)`

	_, err := m.db.Exec(ctx, query,
		message.ID,
		message.TwilioSID,
		message.From,
		message.To,
		message.Direction,
		message.Type,
		message.Status,
		message.Content,
		message.MediaURL,
		message.MediaType,
		message.Timestamp,
		message.CreatedAt,
		message.UpdatedAt,
		message.UserID,
		message.SessionID,
		message.ErrorCode,
		message.ErrorMsg,
	)

	if err != nil {
		m.logger.WithError(err).Error("Failed to store message in database")
		return fmt.Errorf("failed to store message: %w", err)
	}

	// Cache recent messages in Redis for quick access
	cacheKey := fmt.Sprintf("message:%s", message.ID)
	if err := m.redis.Set(ctx, cacheKey, message, 24*time.Hour).Err(); err != nil {
		m.logger.WithError(err).Warn("Failed to cache message in Redis")
	}

	m.logger.WithField("message_id", message.ID).Info("Message stored successfully")
	return nil
}

// GetMessage retrieves a message by ID
func (m *MessageService) GetMessage(ctx context.Context, messageID string) (*models.WhatsAppMessage, error) {
	m.logger.WithField("message_id", messageID).Info("Retrieving message")

	// Parse UUID
	id, err := uuid.Parse(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID format: %w", err)
	}

	// Try cache first
	cacheKey := fmt.Sprintf("message:%s", messageID)
	var message models.WhatsAppMessage
	if err := m.redis.Get(ctx, cacheKey).Scan(&message); err == nil {
		m.logger.WithField("message_id", messageID).Debug("Message retrieved from cache")
		return &message, nil
	}

	// Query database
	query := `
		SELECT id, twilio_sid, from_number, to_number, direction, message_type,
			   status, content, media_url, media_type, timestamp, created_at, updated_at,
			   user_id, session_id, error_code, error_message
		FROM whatsapp_messages 
		WHERE id = $1`

	row := m.db.QueryRow(ctx, query, id)
	
	err = row.Scan(
		&message.ID,
		&message.TwilioSID,
		&message.From,
		&message.To,
		&message.Direction,
		&message.Type,
		&message.Status,
		&message.Content,
		&message.MediaURL,
		&message.MediaType,
		&message.Timestamp,
		&message.CreatedAt,
		&message.UpdatedAt,
		&message.UserID,
		&message.SessionID,
		&message.ErrorCode,
		&message.ErrorMsg,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		m.logger.WithError(err).Error("Failed to retrieve message from database")
		return nil, fmt.Errorf("failed to retrieve message: %w", err)
	}

	// Cache the result
	if err := m.redis.Set(ctx, cacheKey, &message, 24*time.Hour).Err(); err != nil {
		m.logger.WithError(err).Warn("Failed to cache retrieved message")
	}

	m.logger.WithField("message_id", messageID).Info("Message retrieved successfully")
	return &message, nil
}

// UpdateMessageStatus updates the status of a message
func (m *MessageService) UpdateMessageStatus(ctx context.Context, statusUpdate *models.MessageStatusUpdate) error {
	m.logger.WithFields(logrus.Fields{
		"message_sid": statusUpdate.MessageSid,
		"status":      statusUpdate.Status,
		"error_code":  statusUpdate.ErrorCode,
	}).Info("Updating message status")

	query := `
		UPDATE whatsapp_messages 
		SET status = $2, error_code = $3, error_message = $4, updated_at = $5
		WHERE twilio_sid = $1`

	result, err := m.db.Exec(ctx, query,
		statusUpdate.MessageSid,
		statusUpdate.Status,
		statusUpdate.ErrorCode,
		statusUpdate.ErrorMessage,
		statusUpdate.Timestamp,
	)

	if err != nil {
		m.logger.WithError(err).Error("Failed to update message status in database")
		return fmt.Errorf("failed to update message status: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		m.logger.WithField("message_sid", statusUpdate.MessageSid).Warn("No message found to update")
		return fmt.Errorf("message not found for status update")
	}

	// Invalidate cache
	// We don't have the message ID here, so we'll need to query for it or use a different cache strategy
	
	m.logger.WithFields(logrus.Fields{
		"message_sid":   statusUpdate.MessageSid,
		"rows_affected": rowsAffected,
	}).Info("Message status updated successfully")

	return nil
}

// GetMessagesByUser retrieves messages for a specific user/phone number
func (m *MessageService) GetMessagesByUser(ctx context.Context, phoneNumber string, limit int, offset int) ([]*models.WhatsAppMessage, error) {
	m.logger.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"limit":        limit,
		"offset":       offset,
	}).Info("Retrieving messages by user")

	query := `
		SELECT id, twilio_sid, from_number, to_number, direction, message_type,
			   status, content, media_url, media_type, timestamp, created_at, updated_at,
			   user_id, session_id, error_code, error_message
		FROM whatsapp_messages 
		WHERE from_number = $1 OR to_number = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3`

	rows, err := m.db.Query(ctx, query, phoneNumber, limit, offset)
	if err != nil {
		m.logger.WithError(err).Error("Failed to query messages by user")
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.WhatsAppMessage
	for rows.Next() {
		var message models.WhatsAppMessage
		err := rows.Scan(
			&message.ID,
			&message.TwilioSID,
			&message.From,
			&message.To,
			&message.Direction,
			&message.Type,
			&message.Status,
			&message.Content,
			&message.MediaURL,
			&message.MediaType,
			&message.Timestamp,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.UserID,
			&message.SessionID,
			&message.ErrorCode,
			&message.ErrorMsg,
		)
		if err != nil {
			m.logger.WithError(err).Error("Failed to scan message row")
			continue
		}
		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		m.logger.WithError(err).Error("Error iterating over message rows")
		return nil, fmt.Errorf("error reading messages: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"phone_number":   phoneNumber,
		"messages_found": len(messages),
	}).Info("Messages retrieved successfully")

	return messages, nil
}

// GetRecentMessages retrieves recent messages across all users
func (m *MessageService) GetRecentMessages(ctx context.Context, limit int) ([]*models.WhatsAppMessage, error) {
	m.logger.WithField("limit", limit).Info("Retrieving recent messages")

	query := `
		SELECT id, twilio_sid, from_number, to_number, direction, message_type,
			   status, content, media_url, media_type, timestamp, created_at, updated_at,
			   user_id, session_id, error_code, error_message
		FROM whatsapp_messages 
		ORDER BY timestamp DESC
		LIMIT $1`

	rows, err := m.db.Query(ctx, query, limit)
	if err != nil {
		m.logger.WithError(err).Error("Failed to query recent messages")
		return nil, fmt.Errorf("failed to query recent messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.WhatsAppMessage
	for rows.Next() {
		var message models.WhatsAppMessage
		err := rows.Scan(
			&message.ID,
			&message.TwilioSID,
			&message.From,
			&message.To,
			&message.Direction,
			&message.Type,
			&message.Status,
			&message.Content,
			&message.MediaURL,
			&message.MediaType,
			&message.Timestamp,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.UserID,
			&message.SessionID,
			&message.ErrorCode,
			&message.ErrorMsg,
		)
		if err != nil {
			m.logger.WithError(err).Error("Failed to scan message row")
			continue
		}
		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		m.logger.WithError(err).Error("Error iterating over recent message rows")
		return nil, fmt.Errorf("error reading recent messages: %w", err)
	}

	m.logger.WithField("messages_found", len(messages)).Info("Recent messages retrieved successfully")
	return messages, nil
}