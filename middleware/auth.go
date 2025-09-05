package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides basic authentication middleware
// For this chatbot application, we'll implement a simple token-based auth
type AuthMiddleware struct {
	// In a real application, you would validate tokens against a database
	// or JWT service. For now, we'll use a simple approach.
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// RequireAuth middleware that requires authentication
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header is required",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token is required",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Validate the token (simplified for demo purposes)
		if !a.validateToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Set user info in context (you would typically decode JWT or lookup user)
		c.Set("user_id", "demo_user")
		c.Set("token", token)

		c.Next()
	}
}

// OptionalAuth middleware that optionally validates authentication
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if a.validateToken(token) {
				c.Set("user_id", "demo_user")
				c.Set("token", token)
			}
		}
		c.Next()
	}
}

// validateToken validates the provided token
// In a real application, this would validate against a database or JWT
func (a *AuthMiddleware) validateToken(token string) bool {
	// For demo purposes, accept any non-empty token
	// In production, you would:
	// 1. Validate JWT signature
	// 2. Check token expiration
	// 3. Verify token against database
	// 4. Check user permissions

	// Simple validation for demo
	return len(token) > 10
}

// RateLimitMiddleware provides basic rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	// This is a simplified rate limiter
	// In production, you would use Redis or a proper rate limiting library
	return func(c *gin.Context) {
		// For now, just pass through
		// You could implement IP-based rate limiting here
		c.Next()
	}
}

// LoggingMiddleware provides request logging
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
