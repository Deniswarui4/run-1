package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/warui/event-ticketing-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error connecting to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Creating test data...")
	fmt.Println()

	// Find or create a test user (organizer)
	var organizer models.User
	err = db.Where("email = ?", "test-organizer@example.com").First(&organizer).Error
	if err != nil {
		// Create test organizer
		organizer = models.User{
			Email:      "test-organizer@example.com",
			Password:   "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890", // hashed "password123"
			FirstName:  "Test",
			LastName:   "Organizer",
			Role:       "organizer",
			IsActive:   true,
			IsVerified: true,
		}
		if err := db.Create(&organizer).Error; err != nil {
			fmt.Printf("Error creating organizer: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created test organizer: %s\n", organizer.Email)

		// Create organizer balance
		balance := models.OrganizerBalance{
			OrganizerID:      organizer.ID,
			TotalEarnings:    0,
			AvailableBalance: 0,
			PendingBalance:   0,
			WithdrawnAmount:  0,
		}
		db.Create(&balance)
	} else {
		fmt.Printf("✓ Using existing organizer: %s\n", organizer.Email)
	}

	// Create a test event
	var event models.Event
	err = db.Where("title = ? AND organizer_id = ?", "Test Event for Guest Checkout", organizer.ID).First(&event).Error
	if err != nil {
		event = models.Event{
			Title:        "Test Event for Guest Checkout",
			Description:  "This is a test event for testing guest checkout functionality",
			Category:     "Technology",
			Venue:        "Test Venue",
			Address:      "123 Test Street",
			City:         "Nairobi",
			Country:      "Kenya",
			StartDate:    time.Now().Add(30 * 24 * time.Hour), // 30 days from now
			EndDate:      time.Now().Add(31 * 24 * time.Hour), // 31 days from now
			Status:       models.EventStatusPublished,
			IsFeatured:   false,
			FeaturedType: models.FeaturedTypeNone,
			OrganizerID:  organizer.ID,
		}
		if err := db.Create(&event).Error; err != nil {
			fmt.Printf("Error creating event: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created test event: %s (ID: %s)\n", event.Title, event.ID)
	} else {
		fmt.Printf("✓ Using existing event: %s (ID: %s)\n", event.Title, event.ID)
	}

	// Create ticket types
	var ticketTypes []models.TicketType
	db.Where("event_id = ?", event.ID).Find(&ticketTypes)

	if len(ticketTypes) == 0 {
		// Create General Admission ticket
		generalTicket := models.TicketType{
			EventID:     event.ID,
			Name:        "General Admission",
			Description: "Standard entry ticket",
			Price:       1000.0,
			Quantity:    100,
			Sold:        0,
			MaxPerOrder: 10,
			SaleStart:   time.Now(),
			SaleEnd:     event.StartDate,
			IsActive:    true,
		}
		if err := db.Create(&generalTicket).Error; err != nil {
			fmt.Printf("Error creating general ticket: %v\n", err)
		} else {
			fmt.Printf("✓ Created ticket type: %s (ID: %s, Price: %.2f)\n", generalTicket.Name, generalTicket.ID, generalTicket.Price)
		}

		// Create VIP ticket
		vipTicket := models.TicketType{
			EventID:     event.ID,
			Name:        "VIP",
			Description: "VIP access with premium benefits",
			Price:       2500.0,
			Quantity:    50,
			Sold:        0,
			MaxPerOrder: 5,
			SaleStart:   time.Now(),
			SaleEnd:     event.StartDate,
			IsActive:    true,
		}
		if err := db.Create(&vipTicket).Error; err != nil {
			fmt.Printf("Error creating VIP ticket: %v\n", err)
		} else {
			fmt.Printf("✓ Created ticket type: %s (ID: %s, Price: %.2f)\n", vipTicket.Name, vipTicket.ID, vipTicket.Price)
		}
	} else {
		fmt.Printf("✓ Event already has %d ticket type(s)\n", len(ticketTypes))
	}

	fmt.Println()
	fmt.Println("✓ Test data setup complete!")
	fmt.Println()
	fmt.Println("You can now:")
	fmt.Println("1. Test guest checkout with the event above")
	fmt.Println("2. Run: go run ./cmd/test-cart-api/main.go")
	fmt.Println()
	fmt.Printf("Event ID: %s\n", event.ID)
	fmt.Printf("Event URL: http://localhost:3000/events/%s\n", event.ID)
}

func openDB() (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	name := getEnv("DB_NAME", "event_ticketing")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, name, sslmode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to reach database at %s:%s — %w", host, port, err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
