package main

import (
	"chatbot_backend/config"
	"chatbot_backend/handlers"
	"chatbot_backend/middleware"
	"chatbot_backend/models"
	"chatbot_backend/services"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db := initDB()

	// Initialize AI service
	aiService := initAIService(cfg)

	// Initialize router
	r := setupRouter(cfg, db, aiService)

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// initDB initializes the database connection and runs migrations
func initDB() *gorm.DB {
	// Get database configuration from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "chatbot")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&models.Session{}, &models.Message{}, &models.Reaction{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database initialized successfully")
	return db
}

// initAIService initializes the AI service
func initAIService(cfg *config.Config) services.AIService {
	if cfg.AIAPIKey == "" {
		log.Println("No AI API key provided, using mock service")
		return services.NewMockAIService()
	}

	log.Println("Initializing OpenAI service")
	return services.NewOpenAIService()
}

// setupRouter configures and returns the Gin router
func setupRouter(cfg *config.Config, db *gorm.DB, aiService services.AIService) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Add middleware
	r.Use(middleware.LoggingMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"service":     "chatbot-backend",
			"version":     "1.0.0",
			"environment": cfg.Environment,
		})
	})

	// Setup API routes
	setupRoutes(r, db, aiService)

	return r
}

// setupRoutes configures all API routes
func setupRoutes(r *gin.Engine, db *gorm.DB, aiService services.AIService) {
	api := r.Group("/api")

	// Chat routes
	chat := api.Group("/chat")
	chat.POST("/send", handlers.SendMessage(db, aiService))
	chat.POST("/regenerate", handlers.RegenerateMessage(db, aiService))
	chat.GET("/messages/:id", handlers.GetMessages(db))

	// Session routes
	sessions := api.Group("/sessions")
	sessions.GET("", handlers.GetSessions(db))
	sessions.POST("", handlers.CreateSession(db))
	sessions.GET("/:id", handlers.GetSession(db))
	sessions.PUT("/:id", handlers.UpdateSession(db))
	sessions.DELETE("/:id", handlers.DeleteSession(db))
	sessions.POST("/:id/favorite", handlers.ToggleFavorite(db))

	// WebSocket endpoint (placeholder for future implementation)
	r.GET("/ws/chat/:sessionId", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "WebSocket endpoint - not implemented yet",
			"sessionId": c.Param("sessionId"),
		})
	})
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
