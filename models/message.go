package models

import (
	"time"
)

// Message represents a chat message
type Message struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	Content           string     `json:"content"`
	Sender            string     `json:"sender"` // "user" | "bot"
	Timestamp         time.Time  `json:"timestamp"`
	MessageType       string     `json:"messageType"` // "text" | "code" | "image" | "link"
	IsTyping          bool       `json:"isTyping"`
	IsFavorite        bool       `json:"isFavorite"`
	IsRegenerated     bool       `json:"isRegenerated"`
	OriginalMessageID string     `json:"originalMessageId,omitempty"`
	SessionID         string     `json:"sessionId"`
	Reactions         []Reaction `json:"reactions" gorm:"foreignKey:MessageID"`
}

// Reaction represents a message reaction
type Reaction struct {
	ID        string `json:"id" gorm:"primaryKey"`
	Emoji     string `json:"emoji"`
	Count     int    `json:"count"`
	Users     string `json:"users"` // JSON array as string
	MessageID string `json:"messageId"`
}
