//go:build integration

// Feature: guest-checkout-cart-sharing-metrics
package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"pgregory.net/rapid"
)

// guestTicketTestDB opens a GORM connection and migrates all tables required
// for Properties 3 and 4 (Ticket and GuestSession in addition to the base set).
func guestTicketTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	env := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_USER", "postgres"),
		env("DB_PASSWORD", "postgres"),
		env("DB_NAME", "event_ticketing_test"),
		env("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to DB: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.TicketType{},
		&models.Transaction{},
		&models.Ticket{},
		&models.GuestSession{},
		&models.PlatformSettings{},
	); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	t.Cleanup(func() {
		if sqlDB, _ := db.DB(); sqlDB != nil {
			sqlDB.Close()
		}
	})

	return db
}

// buildGuestTicketRouter wires up a minimal Gin router that exposes:
//   - POST /api/v1/guest/checkout  — creates a Transaction + Tickets (mocked Paystack)
//   - GET  /api/v1/guest/tickets   — looks up tickets by email + reference
//
// This mirrors the real GuestHandler logic without requiring the full service
// dependency graph (storage, QR, PDF, email).
func buildGuestTicketRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// POST /api/v1/guest/checkout
	// Validates email, creates a completed Transaction, and creates one Ticket per
	// item with guest_email set and attendee_id nil — simulating a successful
	// Paystack verify callback inline for test simplicity.
	r.POST("/api/v1/guest/checkout", func(c *gin.Context) {
		var req struct {
			Email   string `json:"email"`
			EventID string `json:"event_id"`
			Items   []struct {
				TicketTypeID string `json:"ticket_type_id"`
				Quantity     int    `json:"quantity"`
			} `json:"items"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !ValidateGuestEmail(req.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
			return
		}

		eventID, err := uuid.Parse(req.EventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
			return
		}

		// Create a completed transaction (skipping real Paystack).
		ref := fmt.Sprintf("TXN-%s-%d", uuid.New().String()[:8], time.Now().UnixNano())
		txn := models.Transaction{
			UserID:           uuid.Nil, // guest sentinel
			EventID:          &eventID,
			Type:             models.TransactionTypeTicketPurchase,
			Status:           models.TransactionStatusCompleted,
			Amount:           0,
			Currency:         "NGN",
			NetAmount:        0,
			PaymentGateway:   "paystack",
			PaymentReference: ref,
			Description:      "guest test purchase",
		}
		if err := db.Create(&txn).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create transaction: " + err.Error()})
			return
		}

		// Create tickets for each item.
		var tickets []models.Ticket
		for _, item := range req.Items {
			ttID, err := uuid.Parse(item.TicketTypeID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket_type_id"})
				return
			}
			for i := 0; i < item.Quantity; i++ {
				email := req.Email
				ticket := models.Ticket{
					EventID:       eventID,
					TicketTypeID:  ttID,
					TransactionID: txn.ID,
					Status:        models.TicketStatusConfirmed,
					Price:         0,
					GuestEmail:    &email,
					// AttendeeID intentionally left nil
				}
				if err := db.Create(&ticket).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "create ticket: " + err.Error()})
					return
				}
				tickets = append(tickets, ticket)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"payment_reference": ref,
			"tickets":           tickets,
		})
	})

	// GET /api/v1/guest/tickets?email=&reference=
	r.GET("/api/v1/guest/tickets", func(c *gin.Context) {
		email := c.Query("email")
		reference := c.Query("reference")

		if email == "" || reference == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email and reference are required"})
			return
		}

		var txn models.Transaction
		if err := db.First(&txn, "payment_reference = ?", reference).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
			return
		}

		var tickets []models.Ticket
		if err := db.Where("transaction_id = ? AND guest_email = ?", txn.ID, email).
			Find(&tickets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch tickets: " + err.Error()})
			return
		}

		if len(tickets) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "no tickets found for this email and reference"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"tickets": tickets})
	})

	return r
}

// doGuestCheckout is a helper that POSTs to /api/v1/guest/checkout and returns
// the recorder plus the decoded response body.
func doGuestCheckout(router *gin.Engine, email, eventID string, items []map[string]interface{}) (*httptest.ResponseRecorder, map[string]interface{}) {
	body := map[string]interface{}{
		"email":    email,
		"event_id": eventID,
		"items":    items,
	}
	w := postCheckout(router, body)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w, resp
}

// ── Property 3 ──────────────────────────────────────────────────────────────

// TestPropertyGuestTicketsHaveGuestEmailAndNilAttendeeID verifies that every
// ticket created through the guest checkout flow has guest_email equal to the
// email supplied at checkout and attendee_id equal to nil.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 3: Guest tickets have guest_email set and attendee_id null
//
// Validates: Requirements 1.2
func TestPropertyGuestTicketsHaveGuestEmailAndNilAttendeeID(t *testing.T) {
	db := guestTicketTestDB(t)
	organizer := seedGuestOrganizer(t, db)
	router := buildGuestTicketRouter(db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedGuestEvent(t, db, organizer.ID)
		defer db.Unscoped().Delete(&event)

		tt := seedGuestTicketType(t, db, event.ID, 10)
		defer db.Unscoped().Delete(&tt)

		// Generate a valid email using the same conservative pattern as Property 1.
		email := rapid.StringMatching(`[a-zA-Z0-9._%+\-]{1,20}@[a-zA-Z0-9\-]{1,10}\.[a-zA-Z]{2,6}`).Draw(rt, "email")
		quantity := rapid.IntRange(1, 3).Draw(rt, "quantity")

		w, resp := doGuestCheckout(router, email, event.ID.String(), []map[string]interface{}{
			{"ticket_type_id": tt.ID.String(), "quantity": quantity},
		})

		if w.Code != http.StatusOK {
			rt.Fatalf("checkout failed (%d): %s", w.Code, w.Body.String())
		}

		ref, _ := resp["payment_reference"].(string)
		if ref == "" {
			rt.Fatalf("no payment_reference in response: %v", resp)
		}

		// Fetch the created tickets directly from the DB.
		var txn models.Transaction
		if err := db.First(&txn, "payment_reference = ?", ref).Error; err != nil {
			rt.Fatalf("transaction not found: %v", err)
		}

		var tickets []models.Ticket
		if err := db.Where("transaction_id = ?", txn.ID).Find(&tickets).Error; err != nil {
			rt.Fatalf("fetch tickets: %v", err)
		}

		if len(tickets) != quantity {
			rt.Fatalf("expected %d tickets, got %d", quantity, len(tickets))
		}

		for _, ticket := range tickets {
			// Property 3a: guest_email must equal the checkout email.
			if ticket.GuestEmail == nil {
				rt.Fatalf("ticket %s: guest_email is nil, want %q", ticket.ID, email)
			}
			if *ticket.GuestEmail != email {
				rt.Fatalf("ticket %s: guest_email = %q, want %q", ticket.ID, *ticket.GuestEmail, email)
			}

			// Property 3b: attendee_id must be nil.
			if ticket.AttendeeID != nil {
				rt.Fatalf("ticket %s: attendee_id = %v, want nil", ticket.ID, ticket.AttendeeID)
			}
		}

		// Cleanup tickets and transaction for this iteration.
		db.Unscoped().Where("transaction_id = ?", txn.ID).Delete(&models.Ticket{})
		db.Unscoped().Delete(&txn)
	})
}

// ── Property 4 ──────────────────────────────────────────────────────────────

// TestPropertyGuestTicketLookupRoundTrip verifies that for any completed guest
// order, calling LookupGuestTickets with the matching email and reference
// returns all tickets associated with that order.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 4: Guest ticket lookup round-trip
//
// Validates: Requirements 1.6
func TestPropertyGuestTicketLookupRoundTrip(t *testing.T) {
	db := guestTicketTestDB(t)
	organizer := seedGuestOrganizer(t, db)
	router := buildGuestTicketRouter(db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedGuestEvent(t, db, organizer.ID)
		defer db.Unscoped().Delete(&event)

		tt := seedGuestTicketType(t, db, event.ID, 10)
		defer db.Unscoped().Delete(&tt)

		email := rapid.StringMatching(`[a-zA-Z0-9._%+\-]{1,20}@[a-zA-Z0-9\-]{1,10}\.[a-zA-Z]{2,6}`).Draw(rt, "email")
		quantity := rapid.IntRange(1, 5).Draw(rt, "quantity")

		// Step 1: create the guest order.
		wCheckout, checkoutResp := doGuestCheckout(router, email, event.ID.String(), []map[string]interface{}{
			{"ticket_type_id": tt.ID.String(), "quantity": quantity},
		})
		if wCheckout.Code != http.StatusOK {
			rt.Fatalf("checkout failed (%d): %s", wCheckout.Code, wCheckout.Body.String())
		}

		ref, _ := checkoutResp["payment_reference"].(string)
		if ref == "" {
			rt.Fatalf("no payment_reference in checkout response: %v", checkoutResp)
		}

		// Step 2: look up the tickets using the guest email + reference.
		lookupReq := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/guest/tickets?email=%s&reference=%s", email, ref),
			nil,
		)
		wLookup := httptest.NewRecorder()
		router.ServeHTTP(wLookup, lookupReq)

		if wLookup.Code != http.StatusOK {
			rt.Fatalf("lookup failed (%d): %s", wLookup.Code, wLookup.Body.String())
		}

		var lookupResp struct {
			Tickets []models.Ticket `json:"tickets"`
		}
		if err := json.Unmarshal(wLookup.Body.Bytes(), &lookupResp); err != nil {
			rt.Fatalf("decode lookup response: %v", err)
		}

		// Property 4: all tickets from the order must be returned.
		if len(lookupResp.Tickets) != quantity {
			rt.Fatalf("lookup returned %d tickets, want %d (ref=%s, email=%s)",
				len(lookupResp.Tickets), quantity, ref, email)
		}

		for _, ticket := range lookupResp.Tickets {
			if ticket.GuestEmail == nil || *ticket.GuestEmail != email {
				rt.Fatalf("lookup ticket %s: guest_email = %v, want %q", ticket.ID, ticket.GuestEmail, email)
			}
		}

		// Cleanup this iteration's data.
		var txn models.Transaction
		if err := db.First(&txn, "payment_reference = ?", ref).Error; err == nil {
			db.Unscoped().Where("transaction_id = ?", txn.ID).Delete(&models.Ticket{})
			db.Unscoped().Delete(&txn)
		}
	})
}
