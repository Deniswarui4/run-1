package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/database"
	"github.com/warui/event-ticketing-api/internal/middleware"
	"github.com/warui/event-ticketing-api/internal/routes"
	"github.com/warui/event-ticketing-api/internal/services"
)

// verifySchema checks the goose_db_version table and logs a warning if migrations
// are absent or pending. It never exits — startup always continues.
func verifySchema(sqlDB *sql.DB) {
	var exists bool
	err := sqlDB.QueryRow(
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'goose_db_version'
		)`,
	).Scan(&exists)
	if err != nil || !exists {
		log.Println("WARNING: schema may be out of date — run migrate up")
		return
	}

	var pending int
	err = sqlDB.QueryRow(
		`SELECT COUNT(*) FROM goose_db_version WHERE is_applied = false`,
	).Scan(&pending)
	if err != nil {
		log.Printf("WARNING: could not query goose_db_version: %v", err)
		return
	}
	if pending > 0 {
		log.Println("WARNING: schema has pending migrations — run migrate up")
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration...
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Check schema version and warn if migrations are pending
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("WARNING: could not obtain sql.DB for schema check: %v", err)
	} else {
		verifySchema(sqlDB)
	}

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize router
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Event Ticketing API is running",
		})
	})

	// Initialize routes
	routes.SetupRoutes(router, db, cfg)

	// Start hourly auto-feature evaluation scheduler
	featureService := services.NewFeatureService(db)
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := featureService.EvaluateAutoFeature(); err != nil {
				log.Printf("EvaluateAutoFeature error: %v", err)
			}
		}
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
