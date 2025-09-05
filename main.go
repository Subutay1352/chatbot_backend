package main

import (
	"chatbot_backend/config"
	"chatbot_backend/handlers"
	"chatbot_backend/middleware"
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

	// Check and create tables if they don't exist
	createTablesIfNotExist(db)

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

// createTablesIfNotExist creates tables if they don't exist
func createTablesIfNotExist(db *gorm.DB) {
	// Check if sessions table exists
	if !db.Migrator().HasTable("sessions") {
		log.Println("Creating sessions table...")
		if err := db.Exec(`
			CREATE TABLE sessions (
				id VARCHAR(255) PRIMARY KEY,
				title VARCHAR(255) NOT NULL,
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				is_favorite BOOLEAN DEFAULT FALSE
			)
		`).Error; err != nil {
			log.Fatal("Failed to create sessions table:", err)
		}
		log.Println("Sessions table created successfully")
	}

	// Check if messages table exists
	if !db.Migrator().HasTable("messages") {
		log.Println("Creating messages table...")
		if err := db.Exec(`
			CREATE TABLE messages (
				id VARCHAR(255) PRIMARY KEY,
				content TEXT NOT NULL,
				sender VARCHAR(50) NOT NULL CHECK (sender IN ('user', 'bot')),
				timestamp TIMESTAMP NOT NULL,
				message_type VARCHAR(50) DEFAULT 'text',
				is_typing BOOLEAN DEFAULT FALSE,
				is_favorite BOOLEAN DEFAULT FALSE,
				is_regenerated BOOLEAN DEFAULT FALSE,
				original_message_id VARCHAR(255),
				session_id VARCHAR(255) NOT NULL,
				language VARCHAR(10),
				code_block BOOLEAN DEFAULT FALSE,
				link_title VARCHAR(255),
				link_description TEXT,
				link_image VARCHAR(500),
				link_url VARCHAR(500),
				link_domain VARCHAR(255),
				FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
			)
		`).Error; err != nil {
			log.Fatal("Failed to create messages table:", err)
		}
		log.Println("Messages table created successfully")
	}

	// Check if reactions table exists
	if !db.Migrator().HasTable("reactions") {
		log.Println("Creating reactions table...")
		if err := db.Exec(`
			CREATE TABLE reactions (
				id VARCHAR(255) PRIMARY KEY,
				emoji VARCHAR(10) NOT NULL,
				count INTEGER DEFAULT 0,
				users TEXT,
				message_id VARCHAR(255) NOT NULL,
				FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
			)
		`).Error; err != nil {
			log.Fatal("Failed to create reactions table:", err)
		}
		log.Println("Reactions table created successfully")
	}

	// Create indexes if they don't exist
	createIndexesIfNotExist(db)
}

// createIndexesIfNotExist creates indexes for better performance
func createIndexesIfNotExist(db *gorm.DB) {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_is_favorite ON sessions(is_favorite)",
		"CREATE INDEX IF NOT EXISTS idx_reactions_message_id ON reactions(message_id)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}
	log.Println("Database indexes created/verified")
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
