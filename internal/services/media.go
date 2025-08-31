package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	appConfig "github.com/re9-ai/re9ai-whatsapp-adapter/internal/config"
	"github.com/re9-ai/re9ai-whatsapp-adapter/internal/models"
)

// MediaService handles media file operations and storage
type MediaService struct {
	s3Client *s3.Client
	config   *appConfig.Config
	logger   *logrus.Logger
	bucket   string
}

// NewMediaService creates a new media service instance
func NewMediaService(cfg *appConfig.Config, logger *logrus.Logger) (*MediaService, error) {
	// Load AWS configuration
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	return &MediaService{
		s3Client: s3Client,
		config:   cfg,
		logger:   logger,
		bucket:   cfg.S3BucketName,
	}, nil
}

// UploadMedia uploads a media file to S3 and returns the public URL
func (m *MediaService) UploadMedia(ctx context.Context, file io.Reader, filename, contentType string) (string, error) {
	m.logger.WithFields(logrus.Fields{
		"filename":     filename,
		"content_type": contentType,
	}).Info("Uploading media file to S3")

	// Generate unique key for the file
	fileExt := filepath.Ext(filename)
	fileKey := fmt.Sprintf("whatsapp-media/%s/%s%s", 
		time.Now().Format("2006/01/02"), 
		uuid.New().String(), 
		fileExt,
	)

	// Read file content into buffer
	var buf bytes.Buffer
	_, err := io.Copy(&buf, file)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Upload to S3
	_, err = m.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(m.bucket),
		Key:         aws.String(fileKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
		ACL:         "public-read", // Make file publicly accessible
	})

	if err != nil {
		m.logger.WithError(err).Error("Failed to upload file to S3")
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Construct public URL
	mediaURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", 
		m.bucket, 
		m.config.AWSRegion, 
		fileKey,
	)

	m.logger.WithFields(logrus.Fields{
		"file_key":  fileKey,
		"media_url": mediaURL,
	}).Info("Media file uploaded successfully")

	return mediaURL, nil
}

// ProcessMedia downloads and processes media files from incoming messages
func (m *MediaService) ProcessMedia(ctx context.Context, message *models.WhatsAppMessage) error {
	if message.MediaURL == nil {
		return fmt.Errorf("no media URL provided")
	}

	m.logger.WithFields(logrus.Fields{
		"message_id": message.ID,
		"media_url":  *message.MediaURL,
		"media_type": *message.MediaType,
	}).Info("Processing incoming media")

	// For now, we'll just log the media processing
	// In a full implementation, you might:
	// 1. Download the media from Twilio's URL
	// 2. Perform virus scanning
	// 3. Extract metadata (image dimensions, video duration, etc.)
	// 4. Generate thumbnails for images/videos
	// 5. Store in your own S3 bucket for long-term storage
	// 6. Run AI analysis (image recognition, OCR, etc.)

	switch {
	case strings.HasPrefix(*message.MediaType, "image/"):
		return m.processImage(ctx, message)
	case strings.HasPrefix(*message.MediaType, "video/"):
		return m.processVideo(ctx, message)
	case strings.HasPrefix(*message.MediaType, "audio/"):
		return m.processAudio(ctx, message)
	case strings.HasPrefix(*message.MediaType, "application/pdf"):
		return m.processDocument(ctx, message)
	default:
		m.logger.WithField("media_type", *message.MediaType).Info("Unknown media type, skipping processing")
		return nil
	}
}

// processImage handles image file processing
func (m *MediaService) processImage(ctx context.Context, message *models.WhatsAppMessage) error {
	m.logger.WithField("message_id", message.ID).Info("Processing image file")

	// TODO: Implement image processing logic
	// - Download image from Twilio URL
	// - Generate thumbnails
	// - Extract EXIF data
	// - Run image recognition/OCR
	// - Store processed results

	return nil
}

// processVideo handles video file processing
func (m *MediaService) processVideo(ctx context.Context, message *models.WhatsAppMessage) error {
	m.logger.WithField("message_id", message.ID).Info("Processing video file")

	// TODO: Implement video processing logic
	// - Download video from Twilio URL
	// - Extract thumbnail frames
	// - Get video metadata (duration, resolution, etc.)
	// - Compress if needed
	// - Store processed results

	return nil
}

// processAudio handles audio file processing
func (m *MediaService) processAudio(ctx context.Context, message *models.WhatsAppMessage) error {
	m.logger.WithField("message_id", message.ID).Info("Processing audio file")

	// TODO: Implement audio processing logic
	// - Download audio from Twilio URL
	// - Convert to standard format if needed
	// - Extract audio metadata
	// - Run speech-to-text conversion
	// - Store processed results

	return nil
}

// processDocument handles document file processing
func (m *MediaService) processDocument(ctx context.Context, message *models.WhatsAppMessage) error {
	m.logger.WithField("message_id", message.ID).Info("Processing document file")

	// TODO: Implement document processing logic
	// - Download document from Twilio URL
	// - Extract text content (OCR for images, text extraction for PDFs)
	// - Generate preview/thumbnail
	// - Store processed results

	return nil
}

// GetMediaInfo retrieves metadata about a media file
func (m *MediaService) GetMediaInfo(ctx context.Context, mediaURL string) (map[string]interface{}, error) {
	m.logger.WithField("media_url", mediaURL).Info("Getting media info")

	// TODO: Implement media info extraction
	// This would typically involve:
	// - Downloading the file header
	// - Extracting metadata without downloading the full file
	// - Returning information like file size, dimensions, duration, etc.

	return map[string]interface{}{
		"url":       mediaURL,
		"processed": false,
	}, nil
}

// DeleteMedia removes a media file from storage
func (m *MediaService) DeleteMedia(ctx context.Context, mediaURL string) error {
	m.logger.WithField("media_url", mediaURL).Info("Deleting media file")

	// Extract key from URL
	// This assumes the URL follows the pattern: https://bucket.s3.region.amazonaws.com/key
	parts := strings.Split(mediaURL, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid media URL format")
	}

	// The key is everything after the domain
	key := strings.Join(parts[3:], "/")

	_, err := m.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		m.logger.WithError(err).Error("Failed to delete media file from S3")
		return fmt.Errorf("failed to delete media: %w", err)
	}

	m.logger.WithField("key", key).Info("Media file deleted successfully")
	return nil
}
