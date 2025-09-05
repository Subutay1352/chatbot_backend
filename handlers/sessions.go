package handlers

import (
	"chatbot_backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateSessionRequest represents the request to create a new session
type CreateSessionRequest struct {
	Title string `json:"title"`
}

// UpdateSessionRequest represents the request to update a session
type UpdateSessionRequest struct {
	Title      string `json:"title,omitempty"`
	IsFavorite *bool  `json:"isFavorite,omitempty"`
}

// GetSessions retrieves all sessions
func GetSessions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessions []models.Session
		if err := db.Order("updated_at DESC").Find(&sessions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to retrieve sessions",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"sessions": sessions})
	}
}

// CreateSession creates a new session
func CreateSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Set default title if not provided
		if req.Title == "" {
			req.Title = "New Chat"
		}

		session := models.Session{
			ID:         uuid.New().String(),
			Title:      req.Title,
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

		c.JSON(http.StatusCreated, gin.H{"session": session})
	}
}

// GetSession retrieves a specific session
func GetSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		var session models.Session
		if err := db.Preload("Messages").
			First(&session, "id = ?", sessionID).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Session not found",
				Message: "The specified session does not exist",
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"session": session})
	}
}

// UpdateSession updates a session
func UpdateSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		var req UpdateSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		var session models.Session
		if err := db.First(&session, "id = ?", sessionID).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Session not found",
				Message: "The specified session does not exist",
				Code:    http.StatusNotFound,
			})
			return
		}

		// Update fields if provided
		if req.Title != "" {
			session.Title = req.Title
		}
		if req.IsFavorite != nil {
			session.IsFavorite = *req.IsFavorite
		}
		session.UpdatedAt = time.Now()

		if err := db.Save(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to update session",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"session": session})
	}
}

// DeleteSession deletes a session
func DeleteSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		// Check if session exists
		var session models.Session
		if err := db.First(&session, "id = ?", sessionID).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Session not found",
				Message: "The specified session does not exist",
				Code:    http.StatusNotFound,
			})
			return
		}

		// Delete associated messages first
		if err := db.Where("session_id = ?", sessionID).Delete(&models.Message{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to delete session messages",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		// Delete session
		if err := db.Delete(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to delete session",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Session deleted successfully"})
	}
}

// ToggleFavorite toggles the favorite status of a session
func ToggleFavorite(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		var session models.Session
		if err := db.First(&session, "id = ?", sessionID).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Session not found",
				Message: "The specified session does not exist",
				Code:    http.StatusNotFound,
			})
			return
		}

		// Toggle favorite status
		session.IsFavorite = !session.IsFavorite
		session.UpdatedAt = time.Now()

		if err := db.Save(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Database error",
				Message: "Failed to update session",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"session": session,
			"message": "Favorite status updated",
		})
	}
}
