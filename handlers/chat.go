package handlers

import (
	"chatbot_backend/models"
	"chatbot_backend/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	Message   string `json:"message" binding:"required"`
	SessionID string `json:"sessionId,omitempty"`
}

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	Message   models.Message `json:"message"`
	SessionID string         `json:"sessionId"`
}

// RegenerateMessageRequest represents the request to regenerate a message
type RegenerateMessageRequest struct {
	MessageID string `json:"messageId" binding:"required"`
	SessionID string `json:"sessionId" binding:"required"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SendMessage handles sending a new message
func SendMessage(db *gorm.DB, aiService services.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Create or get session
		var session models.Session
		if req.SessionID != "" {
			if err := db.Preload("Mode").First(&session, "id = ?", req.SessionID).Error; err != nil {
				c.JSON(http.StatusNotFound, ErrorResponse{
					Error:   "Session not found",
					Message: "The specified session does not exist",
					Code:    http.StatusNotFound,
				})
				return
			}
		} else {
			// Create new session
			session = models.Session{
				ID:         uuid.New().String(),
				Title:      "New Chat",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				IsFavorite: false,
			}

			if err := db.Create(&session).Error; err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "Database error",
					Message: "Failed to create session",
					Code:    http.StatusInternalServerError,
				})
				return
			}
		}

		// Create user message
		userMessage := models.Message{
			ID:          uuid.New().String(),
			Content:     req.Message,
			Sender:      "user",
			Timestamp:   time.Now(),
			MessageType: "text",
			SessionID:   session.ID,
		}

		if err := db.Create(&userMessage).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to save user message",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		// Get AI response
		aiResponse, err := aiService.SendMessage(req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "AI Service error",
				Message: "Failed to get AI response",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		// Create bot message
		botMessage := models.Message{
			ID:          uuid.New().String(),
			Content:     aiResponse,
			Sender:      "bot",
			Timestamp:   time.Now(),
			MessageType: "text",
			SessionID:   session.ID,
		}

		if err := db.Create(&botMessage).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to save bot message",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		// Update session timestamp
		session.UpdatedAt = time.Now()
		db.Save(&session)

		c.JSON(http.StatusOK, SendMessageResponse{
			Message:   botMessage,
			SessionID: session.ID,
		})
	}
}

// RegenerateMessage handles regenerating a bot message
func RegenerateMessage(db *gorm.DB, aiService services.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegenerateMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Get the original message
		var originalMessage models.Message
		if err := db.First(&originalMessage, "id = ?", req.MessageID).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Message not found",
				Message: "The specified message does not exist",
				Code:    http.StatusNotFound,
			})
			return
		}

		// Check if session exists
		var session models.Session
		if err := db.First(&session, "id = ?", req.SessionID).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Session not found",
				Message: "The specified session does not exist",
				Code:    http.StatusNotFound,
			})
			return
		}

		// Get the previous user message
		var userMessage models.Message
		if err := db.Where("session_id = ? AND sender = ? AND timestamp < ?",
			req.SessionID, "user", originalMessage.Timestamp).
			Order("timestamp DESC").First(&userMessage).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User message not found",
				Message: "Could not find the user message to regenerate",
				Code:    http.StatusNotFound,
			})
			return
		}

		// Mark original message as regenerated
		originalMessage.IsRegenerated = true
		originalMessage.OriginalMessageID = originalMessage.ID
		db.Save(&originalMessage)

		// Get new AI response
		aiResponse, err := aiService.RegenerateMessage(userMessage.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "AI Service error",
				Message: "Failed to regenerate message",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		// Create new bot message
		newMessage := models.Message{
			ID:                uuid.New().String(),
			Content:           aiResponse,
			Sender:            "bot",
			Timestamp:         time.Now(),
			MessageType:       "text",
			SessionID:         req.SessionID,
			IsRegenerated:     true,
			OriginalMessageID: req.MessageID,
		}

		if err := db.Create(&newMessage).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to save regenerated message",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": newMessage})
	}
}

// GetMessages retrieves messages for a session
func GetMessages(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		var messages []models.Message
		if err := db.Where("session_id = ?", sessionID).
			Order("timestamp ASC").
			Find(&messages).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to retrieve messages",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"messages": messages})
	}
}
