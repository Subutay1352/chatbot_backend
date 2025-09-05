package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// GetCORSConfig returns the CORS configuration
func GetCORSConfig() cors.Config {
	config := cors.DefaultConfig()

	// Allow specific origins
	config.AllowOrigins = []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost:8080",
		"http://127.0.0.1:8080",
	}

	// Allow all origins in development (remove in production)
	config.AllowAllOrigins = false

	// Allow specific methods
	config.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
		"PATCH",
	}

	// Allow specific headers
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
		"X-CSRF-Token",
	}

	// Expose headers
	config.ExposeHeaders = []string{
		"Content-Length",
		"Content-Type",
	}

	// Allow credentials
	config.AllowCredentials = true

	// Cache preflight requests
	config.MaxAge = 12 * time.Hour

	return config
}

// CORSMiddleware returns a CORS middleware function
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(GetCORSConfig())
}
