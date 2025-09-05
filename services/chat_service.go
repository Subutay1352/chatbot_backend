package services

import (
	"chatbot_backend/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatService handles chat-related business logic
type ChatService struct {
	db *gorm.DB
}

// NewChatService creates a new chat service instance
func NewChatService(db *gorm.DB) *ChatService {
	return &ChatService{db: db}
}

// CreateSession creates a new chat session
func (s *ChatService) CreateSession(title string) (*models.Session, error) {
	session := &models.Session{
		ID:         uuid.New().String(),
		Title:      title,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsFavorite: false,
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (s *ChatService) GetSession(sessionID string) (*models.Session, error) {
	var session models.Session
	if err := s.db.Preload("Messages").
		First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessions retrieves all sessions
func (s *ChatService) GetSessions() ([]models.Session, error) {
	var sessions []models.Session
	if err := s.db.Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// UpdateSession updates a session
func (s *ChatService) UpdateSession(sessionID string, title string, isFavorite *bool) (*models.Session, error) {
	var session models.Session
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	if title != "" {
		session.Title = title
	}
	if isFavorite != nil {
		session.IsFavorite = *isFavorite
	}
	session.UpdatedAt = time.Now()

	if err := s.db.Save(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

// DeleteSession deletes a session and its messages
func (s *ChatService) DeleteSession(sessionID string) error {
	// Delete associated messages first
	if err := s.db.Where("session_id = ?", sessionID).Delete(&models.Message{}).Error; err != nil {
		return err
	}

	// Delete session
	if err := s.db.Where("id = ?", sessionID).Delete(&models.Session{}).Error; err != nil {
		return err
	}

	return nil
}

// AddMessage adds a message to a session
func (s *ChatService) AddMessage(sessionID string, content string, sender string, messageType string) (*models.Message, error) {
	message := &models.Message{
		ID:          uuid.New().String(),
		Content:     content,
		Sender:      sender,
		Timestamp:   time.Now(),
		MessageType: messageType,
		SessionID:   sessionID,
	}

	if err := s.db.Create(message).Error; err != nil {
		return nil, err
	}

	// Update session timestamp
	if err := s.db.Model(&models.Session{}).Where("id = ?", sessionID).
		Update("updated_at", time.Now()).Error; err != nil {
		return nil, err
	}

	return message, nil
}

// GetMessages retrieves messages for a session
func (s *ChatService) GetMessages(sessionID string) ([]models.Message, error) {
	var messages []models.Message
	if err := s.db.Where("session_id = ?", sessionID).
		Order("timestamp ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// ToggleFavorite toggles the favorite status of a session
func (s *ChatService) ToggleFavorite(sessionID string) (*models.Session, error) {
	var session models.Session
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	session.IsFavorite = !session.IsFavorite
	session.UpdatedAt = time.Now()

	if err := s.db.Save(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

// GetFavoriteSessions retrieves only favorite sessions
func (s *ChatService) GetFavoriteSessions() ([]models.Session, error) {
	var sessions []models.Session
	if err := s.db.Where("is_favorite = ?", true).
		Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// SearchSessions searches sessions by title
func (s *ChatService) SearchSessions(query string) ([]models.Session, error) {
	var sessions []models.Session
	if err := s.db.Where("title ILIKE ?", "%"+query+"%").
		Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}
