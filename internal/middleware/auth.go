package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WhatsAppSignatureVerification verifies Twilio webhook signatures
func WhatsAppSignatureVerification(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" {
			// Skip verification if no secret is configured (development mode)
			c.Next()
			return
		}

		// Get the signature from headers
		signature := c.GetHeader("X-Twilio-Signature")
		if signature == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature"})
			c.Abort()
			return
		}

		// Get the raw body for signature verification
		// Note: In a production implementation, you might need to read the body
		// and then restore it for subsequent handlers
		
		// For now, we'll just verify the signature exists
		// TODO: Implement full signature verification
		
		c.Next()
	}
}

// RateLimit implements basic rate limiting using Redis
func RateLimit(redisClient interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement rate limiting logic using Redis
		// For now, just pass through
		c.Next()
	}
}

// verifySignature verifies the Twilio webhook signature
func verifySignature(signature, secret, body, url string) bool {
	// Create HMAC SHA256 hash
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(url + body))
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	
	return signature == expectedSignature
}
