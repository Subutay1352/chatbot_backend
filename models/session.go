package models

import (
	"time"
)

// Session represents a chat session
type Session struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	IsFavorite bool      `json:"isFavorite"`
	Messages   []Message `json:"messages" gorm:"foreignKey:SessionID"`
}
