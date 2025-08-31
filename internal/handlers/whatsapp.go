package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/models"
	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/services"
)

// WhatsAppHandler handles WhatsApp webhook endpoints and API operations
type WhatsAppHandler struct {
	whatsappService *services.WhatsAppService
	messageService  *services.MessageService
	mediaService    *services.MediaService
	aiService       *services.AIService
	logger          *logrus.Logger
}

// NewWhatsAppHandler creates a new WhatsApp handler
func NewWhatsAppHandler(
	whatsappService *services.WhatsAppService,
	messageService *services.MessageService,
	mediaService *services.MediaService,
	aiService *services.AIService,
	logger *logrus.Logger,
) *WhatsAppHandler {
	return &WhatsAppHandler{
		whatsappService: whatsappService,
		messageService:  messageService,
		mediaService:    mediaService,
		aiService:       aiService,
		logger:          logger,
	}
}

// VerifyWebhook handles WhatsApp webhook verification
func (h *WhatsAppHandler) VerifyWebhook(c *gin.Context) {
	// Twilio sends a GET request with verification parameters
	verifyToken := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")
	mode := c.Query("hub.mode")

	h.logger.WithFields(logrus.Fields{
		"mode":         mode,
		"verify_token": verifyToken,
		"challenge":    challenge,
	}).Info("Webhook verification request received")

	// For now, we'll accept any verification request
	// In production, you should verify the token matches your configured webhook secret
	if mode == "subscribe" && challenge != "" {
		h.logger.Info("Webhook verification successful")
		c.String(http.StatusOK, challenge)
		return
	}

	h.logger.Warn("Webhook verification failed")
	c.Status(http.StatusBadRequest)
}

// HandleMessage processes incoming WhatsApp messages
func (h *WhatsAppHandler) HandleMessage(c *gin.Context) {
	var webhookData models.TwilioWebhookRequest
	
	// Bind form data from Twilio webhook
	if err := c.ShouldBind(&webhookData); err != nil {
		h.logger.WithError(err).Error("Failed to parse webhook data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"message_sid": webhookData.MessageSid,
		"from":        webhookData.From,
		"to":          webhookData.To,
		"body":        webhookData.Body,
		"num_media":   webhookData.NumMedia,
	}).Info("Received WhatsApp message webhook")

	// Process the incoming message
	message, err := h.whatsappService.ProcessIncomingMessage(&webhookData)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process incoming message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
		return
	}

	// Store message in database
	if err := h.messageService.StoreMessage(c.Request.Context(), message); err != nil {
		h.logger.WithError(err).Error("Failed to store message in database")
		// Don't return error to Twilio, message was processed successfully
	}

	// Process media if present
	if message.MediaURL != nil {
		go h.processMediaAsync(message)
	}

	// Forward message to chat orchestrator for AI processing
	go h.forwardToOrchestrator(message)

	// Return success to Twilio
	c.Status(http.StatusOK)
}

// HandleStatus processes message status updates from Twilio
func (h *WhatsAppHandler) HandleStatus(c *gin.Context) {
	var webhookData models.TwilioWebhookRequest
	
	if err := c.ShouldBind(&webhookData); err != nil {
		h.logger.WithError(err).Error("Failed to parse status webhook data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"message_sid": webhookData.MessageSid,
		"status":      webhookData.SmsStatus,
		"error_code":  webhookData.ErrorCode,
	}).Info("Received WhatsApp status update webhook")

	// Process the status update
	statusUpdate, err := h.whatsappService.ProcessStatusUpdate(&webhookData)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process status update")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process status update"})
		return
	}

	// Update message status in database
	if err := h.messageService.UpdateMessageStatus(c.Request.Context(), statusUpdate); err != nil {
		h.logger.WithError(err).Error("Failed to update message status in database")
		// Don't return error to Twilio
	}

	c.Status(http.StatusOK)
}

// SendMessage handles API requests to send WhatsApp messages
func (h *WhatsAppHandler) SendMessage(c *gin.Context) {
	var request models.SendMessageRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("Failed to parse send message request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"to":      request.To,
		"type":    request.Type,
		"content": request.Content,
	}).Info("Sending WhatsApp message via API")

	var response *models.SendMessageResponse
	var err error

	// Send message based on type
	switch request.Type {
	case models.MessageTypeText, "":
		response, err = h.whatsappService.SendTextMessage(c.Request.Context(), request.To, request.Content)
	
	case models.MessageTypeImage, models.MessageTypeVideo, models.MessageTypeAudio, models.MessageTypeDocument:
		if request.MediaURL == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Media URL required for media messages"})
			return
		}
		mediaType := ""
		if request.MediaType != nil {
			mediaType = *request.MediaType
		}
		response, err = h.whatsappService.SendMediaMessage(c.Request.Context(), request.To, request.Content, *request.MediaURL, mediaType)
	
	default:
		if request.Template != nil {
			response, err = h.whatsappService.SendTemplateMessage(c.Request.Context(), request.To, *request.Template, request.Variables)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported message type"})
			return
		}
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to send WhatsApp message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Store outbound message in database
	outboundMessage := &models.WhatsAppMessage{
		ID:        response.ID,
		TwilioSID: response.TwilioSID,
		From:      h.whatsappService.fromNumber,
		To:        request.To,
		Direction: models.MessageDirectionOutbound,
		Type:      request.Type,
		Status:    response.Status,
		Content:   request.Content,
		MediaURL:  request.MediaURL,
		MediaType: request.MediaType,
		Timestamp: response.CreatedAt,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.CreatedAt,
	}

	if err := h.messageService.StoreMessage(c.Request.Context(), outboundMessage); err != nil {
		h.logger.WithError(err).Error("Failed to store outbound message")
		// Don't fail the request, message was sent successfully
	}

	c.JSON(http.StatusOK, response)
}

// GetMessage retrieves a message by ID
func (h *WhatsAppHandler) GetMessage(c *gin.Context) {
	messageID := c.Param("messageId")
	
	h.logger.WithField("message_id", messageID).Info("Retrieving message")

	message, err := h.messageService.GetMessage(c.Request.Context(), messageID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to retrieve message")
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// UploadMedia handles media file uploads
func (h *WhatsAppHandler) UploadMedia(c *gin.Context) {
	file, header, err := c.Request.FormFile("media")
	if err != nil {
		h.logger.WithError(err).Error("Failed to get uploaded file")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	h.logger.WithFields(logrus.Fields{
		"filename": header.Filename,
		"size":     header.Size,
		"type":     header.Header.Get("Content-Type"),
	}).Info("Processing media upload")

	// Upload media to storage service
	mediaURL, err := h.mediaService.UploadMedia(c.Request.Context(), file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		h.logger.WithError(err).Error("Failed to upload media")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media_url": mediaURL,
		"filename":  header.Filename,
		"size":      header.Size,
	})
}

// Helper methods for async processing

// processMediaAsync processes media files in the background
func (h *WhatsAppHandler) processMediaAsync(message *models.WhatsAppMessage) {
	if message.MediaURL == nil {
		return
	}

	h.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"media_url":  *message.MediaURL,
		"media_type": *message.MediaType,
	}).Info("Processing media asynchronously")

	// Download and process media
	err := h.mediaService.ProcessMedia(context.Background(), message)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process media")
	}
}

// forwardToOrchestrator forwards the message to the chat orchestrator
func (h *WhatsAppHandler) forwardToOrchestrator(message *models.WhatsAppMessage) {
	h.logger.WithField("message_id", message.ID).Info("Forwarding message to chat orchestrator")

	err := h.aiService.ForwardToOrchestrator(context.Background(), message)
	if err != nil {
		h.logger.WithError(err).Error("Failed to forward message to orchestrator")
	}
}