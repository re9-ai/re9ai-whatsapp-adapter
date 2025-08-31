package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"

	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/config"
	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/models"
)

// WhatsAppService handles WhatsApp message operations via Twilio
type WhatsAppService struct {
	client     *twilio.RestClient
	config     *config.Config
	logger     *logrus.Logger
	fromNumber string
}

// NewWhatsAppService creates a new WhatsApp service instance
func NewWhatsAppService(cfg *config.Config, logger *logrus.Logger) *WhatsAppService {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: cfg.TwilioAccountSID,
		Password: cfg.TwilioAuthToken,
	})

	return &WhatsAppService{
		client:     client,
		config:     cfg,
		logger:     logger,
		fromNumber: cfg.TwilioWhatsAppFrom,
	}
}

// SendTextMessage sends a text message via WhatsApp
func (w *WhatsAppService) SendTextMessage(ctx context.Context, to, content string) (*models.SendMessageResponse, error) {
	w.logger.WithFields(logrus.Fields{
		"to":      to,
		"content": content,
	}).Info("Sending WhatsApp text message")

	// Ensure the 'to' number has WhatsApp prefix
	toNumber := w.formatWhatsAppNumber(to)

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(toNumber)
	params.SetFrom(w.fromNumber)
	params.SetBody(content)

	resp, err := w.client.Api.CreateMessage(params)
	if err != nil {
		w.logger.WithError(err).Error("Failed to send WhatsApp message")
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	response := &models.SendMessageResponse{
		ID:        uuid.New(),
		TwilioSID: *resp.Sid,
		Status:    models.MessageStatusSent,
		CreatedAt: time.Now(),
	}

	w.logger.WithFields(logrus.Fields{
		"twilio_sid": *resp.Sid,
		"status":     *resp.Status,
	}).Info("WhatsApp message sent successfully")

	return response, nil
}

// SendMediaMessage sends a media message via WhatsApp
func (w *WhatsAppService) SendMediaMessage(ctx context.Context, to, content, mediaURL, mediaType string) (*models.SendMessageResponse, error) {
	w.logger.WithFields(logrus.Fields{
		"to":         to,
		"content":    content,
		"media_url":  mediaURL,
		"media_type": mediaType,
	}).Info("Sending WhatsApp media message")

	toNumber := w.formatWhatsAppNumber(to)

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(toNumber)
	params.SetFrom(w.fromNumber)
	
	if content != "" {
		params.SetBody(content)
	}
	
	// Add media URL
	mediaUrls := []string{mediaURL}
	params.SetMediaUrl(mediaUrls)

	resp, err := w.client.Api.CreateMessage(params)
	if err != nil {
		w.logger.WithError(err).Error("Failed to send WhatsApp media message")
		return nil, fmt.Errorf("failed to send media message: %w", err)
	}

	response := &models.SendMessageResponse{
		ID:        uuid.New(),
		TwilioSID: *resp.Sid,
		Status:    models.MessageStatusSent,
		CreatedAt: time.Now(),
	}

	w.logger.WithFields(logrus.Fields{
		"twilio_sid": *resp.Sid,
		"status":     *resp.Status,
	}).Info("WhatsApp media message sent successfully")

	return response, nil
}

// SendTemplateMessage sends a template message with variables
func (w *WhatsAppService) SendTemplateMessage(ctx context.Context, to, templateSID string, variables map[string]string) (*models.SendMessageResponse, error) {
	w.logger.WithFields(logrus.Fields{
		"to":           to,
		"template_sid": templateSID,
		"variables":    variables,
	}).Info("Sending WhatsApp template message")

	toNumber := w.formatWhatsAppNumber(to)

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(toNumber)
	params.SetFrom(w.fromNumber)
	params.SetContentSid(templateSID)

	// Convert variables to Twilio format
	if len(variables) > 0 {
		contentVariables := make(map[string]interface{})
		for k, v := range variables {
			contentVariables[k] = v
		}
		params.SetContentVariables(contentVariables)
	}

	resp, err := w.client.Api.CreateMessage(params)
	if err != nil {
		w.logger.WithError(err).Error("Failed to send WhatsApp template message")
		return nil, fmt.Errorf("failed to send template message: %w", err)
	}

	response := &models.SendMessageResponse{
		ID:        uuid.New(),
		TwilioSID: *resp.Sid,
		Status:    models.MessageStatusSent,
		CreatedAt: time.Now(),
	}

	w.logger.WithFields(logrus.Fields{
		"twilio_sid": *resp.Sid,
		"status":     *resp.Status,
	}).Info("WhatsApp template message sent successfully")

	return response, nil
}

// ProcessIncomingMessage processes an incoming WhatsApp message from Twilio webhook
func (w *WhatsAppService) ProcessIncomingMessage(webhookData *models.TwilioWebhookRequest) (*models.WhatsAppMessage, error) {
	w.logger.WithFields(logrus.Fields{
		"message_sid": webhookData.MessageSid,
		"from":        webhookData.From,
		"to":          webhookData.To,
	}).Info("Processing incoming WhatsApp message")

	// Determine message type based on media presence
	messageType := models.MessageTypeText
	var mediaURL, mediaType *string

	if numMedia, err := strconv.Atoi(webhookData.NumMedia); err == nil && numMedia > 0 {
		if webhookData.MediaUrl0 != "" {
			mediaURL = &webhookData.MediaUrl0
			mediaType = &webhookData.MediaContentType0
			messageType = w.determineMessageType(webhookData.MediaContentType0)
		}
	}

	// Parse timestamp
	timestamp := time.Now()
	if webhookData.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, webhookData.Timestamp); err == nil {
			timestamp = parsed
		}
	}

	message := &models.WhatsAppMessage{
		ID:        uuid.New(),
		TwilioSID: webhookData.MessageSid,
		From:      webhookData.From,
		To:        webhookData.To,
		Direction: models.MessageDirectionInbound,
		Type:      messageType,
		Status:    models.MessageStatusDelivered,
		Content:   webhookData.Body,
		MediaURL:  mediaURL,
		MediaType: mediaType,
		Timestamp: timestamp,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	w.logger.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"message_type": messageType,
		"content_len":  len(webhookData.Body),
	}).Info("Incoming WhatsApp message processed successfully")

	return message, nil
}

// ProcessStatusUpdate processes a message status update from Twilio webhook
func (w *WhatsAppService) ProcessStatusUpdate(webhookData *models.TwilioWebhookRequest) (*models.MessageStatusUpdate, error) {
	w.logger.WithFields(logrus.Fields{
		"message_sid": webhookData.MessageSid,
		"status":      webhookData.SmsStatus,
	}).Info("Processing WhatsApp message status update")

	status := w.mapTwilioStatus(webhookData.SmsStatus)
	
	update := &models.MessageStatusUpdate{
		MessageSid: webhookData.MessageSid,
		Status:     status,
		Timestamp:  time.Now(),
	}

	// Handle error cases
	if webhookData.ErrorCode != "" {
		update.ErrorCode = &webhookData.ErrorCode
		update.ErrorMessage = &webhookData.ErrorMessage
		update.Status = models.MessageStatusFailed
	}

	w.logger.WithFields(logrus.Fields{
		"mapped_status": status,
		"error_code":    webhookData.ErrorCode,
	}).Info("WhatsApp message status update processed")

	return update, nil
}

// GetMessageStatus retrieves the current status of a message from Twilio
func (w *WhatsAppService) GetMessageStatus(ctx context.Context, messageSID string) (models.MessageStatus, error) {
	w.logger.WithField("message_sid", messageSID).Info("Fetching message status from Twilio")

	params := &twilioApi.FetchMessageParams{}
	resp, err := w.client.Api.FetchMessage(messageSID, params)
	if err != nil {
		w.logger.WithError(err).Error("Failed to fetch message status from Twilio")
		return models.MessageStatusFailed, fmt.Errorf("failed to fetch message status: %w", err)
	}

	status := w.mapTwilioStatus(*resp.Status)
	
	w.logger.WithFields(logrus.Fields{
		"twilio_status": *resp.Status,
		"mapped_status": status,
	}).Info("Message status fetched successfully")

	return status, nil
}

// GetFromNumber returns the configured WhatsApp from number
func (w *WhatsAppService) GetFromNumber() string {
	return w.fromNumber
}

// Helper methods

// formatWhatsAppNumber ensures the phone number has the proper WhatsApp prefix
func (w *WhatsAppService) formatWhatsAppNumber(phoneNumber string) string {
	if strings.HasPrefix(phoneNumber, "whatsapp:") {
		return phoneNumber
	}
	
	// Remove any non-digit characters except +
	cleaned := strings.ReplaceAll(phoneNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	
	// Ensure it starts with +
	if !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}
	
	return fmt.Sprintf("whatsapp:%s", cleaned)
}

// determineMessageType determines the message type based on media content type
func (w *WhatsAppService) determineMessageType(contentType string) models.MessageType {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return models.MessageTypeImage
	case strings.HasPrefix(contentType, "video/"):
		return models.MessageTypeVideo
	case strings.HasPrefix(contentType, "audio/"):
		return models.MessageTypeAudio
	case strings.HasPrefix(contentType, "application/pdf"):
		return models.MessageTypeDocument
	case strings.HasPrefix(contentType, "application/"):
		return models.MessageTypeDocument
	default:
		return models.MessageTypeText
	}
}

// mapTwilioStatus maps Twilio status to our internal status
func (w *WhatsAppService) mapTwilioStatus(twilioStatus string) models.MessageStatus {
	switch strings.ToLower(twilioStatus) {
	case "queued", "accepted":
		return models.MessageStatusPending
	case "sent":
		return models.MessageStatusSent
	case "delivered":
		return models.MessageStatusDelivered
	case "read":
		return models.MessageStatusRead
	case "failed", "undelivered":
		return models.MessageStatusFailed
	default:
		return models.MessageStatusPending
	}
}