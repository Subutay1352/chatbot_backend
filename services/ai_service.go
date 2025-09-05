package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AIService interface defines the contract for AI services
type AIService interface {
	SendMessage(message string) (string, error)
	RegenerateMessage(message string) (string, error)
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message represents a message in the OpenAI API format
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response structure from OpenAI API
type OpenAIResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a choice in the OpenAI response
type Choice struct {
	Message Message `json:"message"`
}

// APIError represents an error from the OpenAI API
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// OpenAIService implements the AIService interface using OpenAI API
type OpenAIService struct {
	APIKey string
	APIURL string
	Client *http.Client
}

// NewOpenAIService creates a new OpenAI service instance
func NewOpenAIService() *OpenAIService {
	return &OpenAIService{
		APIKey: os.Getenv("AI_API_KEY"),
		APIURL: os.Getenv("AI_API_URL"),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendMessage sends a message to the AI service and returns the response
func (s *OpenAIService) SendMessage(message string) (string, error) {
	request := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant. Provide clear and useful responses to user questions."},
			{Role: "user", Content: message},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	return s.makeRequest(request)
}

// RegenerateMessage regenerates a response for the given message
func (s *OpenAIService) RegenerateMessage(message string) (string, error) {
	request := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant. Please provide a different perspective or approach to the user's question."},
			{Role: "user", Content: message},
		},
		MaxTokens:   1000,
		Temperature: 0.8, // Slightly higher temperature for more variation
	}

	return s.makeRequest(request)
}

// makeRequest makes an HTTP request to the OpenAI API
func (s *OpenAIService) makeRequest(request OpenAIRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("API error: %s", apiError.Message)
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices received")
	}

	return response.Choices[0].Message.Content, nil
}

// MockAIService is a mock implementation for testing purposes
type MockAIService struct{}

// NewMockAIService creates a new mock AI service
func NewMockAIService() *MockAIService {
	return &MockAIService{}
}

// SendMessage returns a mock response
func (m *MockAIService) SendMessage(message string) (string, error) {
	return fmt.Sprintf("Mock response to: %s", message), nil
}

// RegenerateMessage returns a mock regenerated response
func (m *MockAIService) RegenerateMessage(message string) (string, error) {
	return fmt.Sprintf("Mock regenerated response to: %s", message), nil
}
