package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/config"
	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/models"
)

// AIService handles communication with AI processing services
type AIService struct {
	config            *config.Config
	logger            *logrus.Logger
	httpClient        *http.Client
	orchestratorURL   string
	aiProcessingURL   string
}

// NewAIService creates a new AI service instance
func NewAIService(cfg *config.Config, logger *logrus.Logger) *AIService {
	return &AIService{
		config:          cfg,
		logger:          logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		orchestratorURL: cfg.ChatOrchestratorURL,
		aiProcessingURL: cfg.AIProcessingURL,
	}
}

// ChatRequest represents a request to the chat orchestrator
type ChatRequest struct {
	MessageID   string                 `json:"message_id"`
	UserPhone   string                 `json:"user_phone"`
	Content     string                 `json:"content"`
	MessageType models.MessageType     `json:"message_type"`
	MediaURL    *string               `json:"media_url,omitempty"`
	MediaType   *string               `json:"media_type,omitempty"`
	Timestamp   time.Time             `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse represents a response from the chat orchestrator
type ChatResponse struct {
	ResponseID    string                 `json:"response_id"`
	Content       string                 `json:"content"`
	MessageType   models.MessageType     `json:"message_type"`
	MediaURL      *string               `json:"media_url,omitempty"`
	MediaType     *string               `json:"media_type,omitempty"`
	ShouldReply   bool                  `json:"should_reply"`
	Context       map[string]interface{} `json:"context,omitempty"`
	NextAction    string                `json:"next_action,omitempty"`
	ProcessedAt   time.Time             `json:"processed_at"`
}

// ForwardToOrchestrator forwards a message to the chat orchestrator for AI processing
func (a *AIService) ForwardToOrchestrator(ctx context.Context, message *models.WhatsAppMessage) error {
	a.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"from":       message.From,
		"content":    message.Content,
	}).Info("Forwarding message to chat orchestrator")

	// Prepare the request payload
	request := ChatRequest{
		MessageID:   message.ID.String(),
		UserPhone:   message.From,
		Content:     message.Content,
		MessageType: message.Type,
		MediaURL:    message.MediaURL,
		MediaType:   message.MediaType,
		Timestamp:   message.Timestamp,
		Context: map[string]interface{}{
			"platform":    "whatsapp",
			"twilio_sid":  message.TwilioSID,
			"direction":   message.Direction,
		},
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		a.logger.WithError(err).Error("Failed to marshal chat request")
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to orchestrator
	url := fmt.Sprintf("%s/api/v1/chat/process", a.orchestratorURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		a.logger.WithError(err).Error("Failed to create HTTP request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "re9ai-whatsapp-adapter/1.0")

	// Make the request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.logger.WithError(err).Error("Failed to send request to orchestrator")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		a.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		}).Error("Orchestrator returned error status")
		return fmt.Errorf("orchestrator returned status %d", resp.StatusCode)
	}

	// Parse response
	var chatResponse ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		a.logger.WithError(err).Error("Failed to decode orchestrator response")
		return fmt.Errorf("failed to decode response: %w", err)
	}

	a.logger.WithFields(logrus.Fields{
		"response_id":   chatResponse.ResponseID,
		"should_reply":  chatResponse.ShouldReply,
		"next_action":   chatResponse.NextAction,
		"content_len":   len(chatResponse.Content),
	}).Info("Received response from chat orchestrator")

	// TODO: Handle the response - this might involve:
	// 1. Sending an automated reply if should_reply is true
	// 2. Triggering additional actions based on next_action
	// 3. Updating user context/session state
	// 4. Logging conversation analytics

	return nil
}

// ProcessDocumentAI sends a document for AI analysis
func (a *AIService) ProcessDocumentAI(ctx context.Context, message *models.WhatsAppMessage, documentURL string) error {
	a.logger.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"document_url": documentURL,
	}).Info("Sending document for AI analysis")

	request := map[string]interface{}{
		"message_id":   message.ID.String(),
		"document_url": documentURL,
		"user_phone":   message.From,
		"context": map[string]interface{}{
			"platform":   "whatsapp",
			"timestamp":  message.Timestamp,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal document AI request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/documents/analyze", a.aiProcessingURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create document AI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send document AI request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("document AI service returned status %d", resp.StatusCode)
	}

	a.logger.WithField("message_id", message.ID).Info("Document sent for AI analysis successfully")
	return nil
}

// ProcessImageAI sends an image for AI analysis
func (a *AIService) ProcessImageAI(ctx context.Context, message *models.WhatsAppMessage, imageURL string) error {
	a.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"image_url":  imageURL,
	}).Info("Sending image for AI analysis")

	request := map[string]interface{}{
		"message_id": message.ID.String(),
		"image_url":  imageURL,
		"user_phone": message.From,
		"context": map[string]interface{}{
			"platform":  "whatsapp",
			"timestamp": message.Timestamp,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal image AI request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/images/analyze", a.aiProcessingURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create image AI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send image AI request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("image AI service returned status %d", resp.StatusCode)
	}

	a.logger.WithField("message_id", message.ID).Info("Image sent for AI analysis successfully")
	return nil
}

// ProcessAudioAI sends audio for speech-to-text processing
func (a *AIService) ProcessAudioAI(ctx context.Context, message *models.WhatsAppMessage, audioURL string) error {
	a.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"audio_url":  audioURL,
	}).Info("Sending audio for speech-to-text processing")

	request := map[string]interface{}{
		"message_id": message.ID.String(),
		"audio_url":  audioURL,
		"user_phone": message.From,
		"context": map[string]interface{}{
			"platform":  "whatsapp",
			"timestamp": message.Timestamp,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal audio AI request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/audio/transcribe", a.aiProcessingURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create audio AI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send audio AI request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("audio AI service returned status %d", resp.StatusCode)
	}

	a.logger.WithField("message_id", message.ID).Info("Audio sent for AI processing successfully")
	return nil
}

// GetConversationContext retrieves conversation context for a user
func (a *AIService) GetConversationContext(ctx context.Context, userPhone string) (map[string]interface{}, error) {
	a.logger.WithField("user_phone", userPhone).Info("Retrieving conversation context")

	url := fmt.Sprintf("%s/api/v1/context/%s", a.orchestratorURL, userPhone)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create context request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation context: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No context found, return empty context
		return map[string]interface{}{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("context service returned status %d", resp.StatusCode)
	}

	var context map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&context); err != nil {
		return nil, fmt.Errorf("failed to decode context response: %w", err)
	}

	a.logger.WithFields(logrus.Fields{
		"user_phone":   userPhone,
		"context_keys": len(context),
	}).Info("Conversation context retrieved successfully")

	return context, nil
}
