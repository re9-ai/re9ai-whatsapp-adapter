package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageDirection represents the direction of a message
type MessageDirection string

const (
	MessageDirectionInbound  MessageDirection = "inbound"
	MessageDirectionOutbound MessageDirection = "outbound"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// MessageType represents the type of message content
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeDocument MessageType = "document"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeLocation MessageType = "location"
	MessageTypeContact  MessageType = "contact"
)

// WhatsAppMessage represents a WhatsApp message in our system
type WhatsAppMessage struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	TwilioSID   string          `json:"twilio_sid" db:"twilio_sid"`
	From        string          `json:"from" db:"from_number"`
	To          string          `json:"to" db:"to_number"`
	Direction   MessageDirection `json:"direction" db:"direction"`
	Type        MessageType      `json:"type" db:"message_type"`
	Status      MessageStatus    `json:"status" db:"status"`
	Content     string          `json:"content" db:"content"`
	MediaURL    *string         `json:"media_url,omitempty" db:"media_url"`
	MediaType   *string         `json:"media_type,omitempty" db:"media_type"`
	Timestamp   time.Time       `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`

	// Additional metadata
	UserID      *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	SessionID   *uuid.UUID `json:"session_id,omitempty" db:"session_id"`
	ErrorCode   *string    `json:"error_code,omitempty" db:"error_code"`
	ErrorMsg    *string    `json:"error_message,omitempty" db:"error_message"`
}

// TwilioWebhookRequest represents incoming webhook payload from Twilio
type TwilioWebhookRequest struct {
	MessageSid          string `form:"MessageSid" json:"MessageSid"`
	AccountSid          string `form:"AccountSid" json:"AccountSid"`
	MessagingServiceSid string `form:"MessagingServiceSid" json:"MessagingServiceSid"`
	From                string `form:"From" json:"From"`
	To                  string `form:"To" json:"To"`
	Body                string `form:"Body" json:"Body"`
	NumMedia            string `form:"NumMedia" json:"NumMedia"`
	MediaContentType0   string `form:"MediaContentType0" json:"MediaContentType0"`
	MediaUrl0           string `form:"MediaUrl0" json:"MediaUrl0"`
	Timestamp           string `form:"Timestamp" json:"Timestamp"`
	ApiVersion          string `form:"ApiVersion" json:"ApiVersion"`
	SmsStatus           string `form:"SmsStatus" json:"SmsStatus"`
	SmsSid              string `form:"SmsSid" json:"SmsSid"`
	SmsMessageSid       string `form:"SmsMessageSid" json:"SmsMessageSid"`
	ErrorCode           string `form:"ErrorCode" json:"ErrorCode"`
	ErrorMessage        string `form:"ErrorMessage" json:"ErrorMessage"`

	// Profile information
	ProfileName string `form:"ProfileName" json:"ProfileName"`
	WaId        string `form:"WaId" json:"WaId"`
}

// SendMessageRequest represents a request to send a WhatsApp message
type SendMessageRequest struct {
	To        string            `json:"to" validate:"required"`
	Content   string            `json:"content" validate:"required"`
	Type      MessageType       `json:"type"`
	MediaURL  *string           `json:"media_url,omitempty"`
	MediaType *string           `json:"media_type,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
	Template  *string           `json:"template,omitempty"`
}

// SendMessageResponse represents the response from sending a message
type SendMessageResponse struct {
	ID        uuid.UUID     `json:"id"`
	TwilioSID string        `json:"twilio_sid"`
	Status    MessageStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

// MessageStatusUpdate represents a status update for a message
type MessageStatusUpdate struct {
	MessageSid   string        `json:"message_sid"`
	Status       MessageStatus `json:"status"`
	ErrorCode    *string       `json:"error_code,omitempty"`
	ErrorMessage *string       `json:"error_message,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// User represents a WhatsApp user in our system
type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	WhatsAppID  string    `json:"whatsapp_id" db:"whatsapp_id"`
	ProfileName string    `json:"profile_name" db:"profile_name"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ChatSession represents a chat conversation session
type ChatSession struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Status    string    `json:"status" db:"status"`
	Context   string    `json:"context" db:"context"`
	StartedAt time.Time `json:"started_at" db:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}