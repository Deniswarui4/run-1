//go:build integration

package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"pgregory.net/rapid"
)

// cartTestDB opens a GORM connection to the test database and auto-migrates
// the tables needed by CartService tests.
func cartTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	getEnv := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "event_ticketing_test"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to DB: %v", err)
	}

	// Auto-migrate all tables required by CartService.
	if err := db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.TicketType{},
		&models.DraftCart{},
		&models.DraftCartItem{},
	); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	return db
}

// seedOrganizer inserts a minimal User row and returns it.
func seedOrganizer(t *testing.T, db *gorm.DB) models.User {
	t.Helper()
	u := models.User{
		ID:        uuid.New(),
		Email:     "organizer-" + uuid.New().String() + "@test.com",
		Password:  "hashed",
		FirstName: "Test",
		LastName:  "Organizer",
		Role:      models.RoleOrganizer,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seedOrganizer: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&u) })
	return u
}

// seedEvent inserts a minimal Event row and returns it.
func seedEvent(t *testing.T, db *gorm.DB, organizerID uuid.UUID) models.Event {
	t.Helper()
	now := time.Now()
	e := models.Event{
		ID:          uuid.New(),
		Title:       "Test Event " + uuid.New().String()[:8],
		Venue:       "Test Venue",
		StartDate:   now.Add(24 * time.Hour),
		EndDate:     now.Add(48 * time.Hour),
		Status:      models.EventStatusPublished,
		OrganizerID: organizerID,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seedEvent: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&e) })
	return e
}

// seedTicketType inserts a TicketType with an active sale window and returns it.
func seedTicketType(t *testing.T, db *gorm.DB, eventID uuid.UUID) models.TicketType {
	t.Helper()
	now := time.Now()
	tt := models.TicketType{
		ID:          uuid.New(),
		EventID:     eventID,
		Name:        "General " + uuid.New().String()[:8],
		Price:       1000,
		Quantity:    100,
		Sold:        0,
		MaxPerOrder: 10,
		SaleStart:   now.Add(-time.Hour),
		SaleEnd:     now.Add(24 * time.Hour),
		IsActive:    true,
	}
	if err := db.Create(&tt).Error; err != nil {
		t.Fatalf("seedTicketType: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&tt) })
	return tt
}

// cleanupCarts removes all draft_cart_items and draft_carts for the given event.
func cleanupCarts(t *testing.T, db *gorm.DB, eventID uuid.UUID) {
	t.Helper()
	db.Exec("DELETE FROM draft_cart_items WHERE cart_id IN (SELECT id FROM draft_carts WHERE event_id = ?)", eventID)
	db.Exec("DELETE FROM draft_carts WHERE event_id = ?", eventID)
}

// ── Property 6 ──────────────────────────────────────────────────────────────

// TestPropertyCartPersistenceRoundTrip verifies that UpsertCart followed by
// GetCart returns a cart with the same ticket type IDs and quantities.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 6: Cart persistence round-trip
//
// Validates: Requirements 2.1, 2.2, 2.3
func TestPropertyCartPersistenceRoundTrip(t *testing.T) {
	db := cartTestDB(t)
	svc := NewCartService(db)

	organizer := seedOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedEvent(t, db, organizer.ID)
		defer cleanupCarts(t, db, event.ID)
		defer db.Unscoped().Delete(&event)

		// Generate 1–3 ticket types for this event.
		numTypes := rapid.IntRange(1, 3).Draw(rt, "numTypes")
		ticketTypes := make([]models.TicketType, numTypes)
		for i := 0; i < numTypes; i++ {
			ticketTypes[i] = seedTicketType(t, db, event.ID)
			defer db.Unscoped().Delete(&ticketTypes[i])
		}

		// Build cart items with random quantities 1–5.
		items := make([]CartItemInput, numTypes)
		for i, tt := range ticketTypes {
			items[i] = CartItemInput{
				TicketTypeID: tt.ID,
				Quantity:     rapid.IntRange(1, 5).Draw(rt, fmt.Sprintf("qty%d", i)),
			}
		}

		// Choose user or guest identity.
		useUser := rapid.Bool().Draw(rt, "useUser")
		var userID *uuid.UUID
		guestSession := ""
		if useUser {
			uid := uuid.New()
			userID = &uid
		} else {
			guestSession = "guest-" + uuid.New().String()
		}

		// Upsert then get.
		_, err := svc.UpsertCart(userID, guestSession, event.ID, items)
		if err != nil {
			rt.Fatalf("UpsertCart failed: %v", err)
		}

		got, err := svc.GetCart(userID, guestSession, event.ID)
		if err != nil {
			rt.Fatalf("GetCart failed: %v", err)
		}
		if got == nil {
			rt.Fatal("GetCart returned nil after UpsertCart")
		}

		// Build expected map: ticketTypeID → quantity.
		expected := make(map[uuid.UUID]int, len(items))
		for _, item := range items {
			expected[item.TicketTypeID] = item.Quantity
		}

		if len(got.Items) != len(expected) {
			rt.Fatalf("item count mismatch: want %d, got %d", len(expected), len(got.Items))
		}
		for _, gotItem := range got.Items {
			wantQty, ok := expected[gotItem.TicketTypeID]
			if !ok {
				rt.Fatalf("unexpected ticket type %s in restored cart", gotItem.TicketTypeID)
			}
			if gotItem.Quantity != wantQty {
				rt.Fatalf("quantity mismatch for %s: want %d, got %d",
					gotItem.TicketTypeID, wantQty, gotItem.Quantity)
			}
		}
	})
}

// ── Property 7 ──────────────────────────────────────────────────────────────

// TestPropertyExpiredCartsNotRestored verifies that a cart whose expires_at is
// in the past is not returned by GetCart.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 7: Expired carts are not restored
//
// Validates: Requirements 2.4
func TestPropertyExpiredCartsNotRestored(t *testing.T) {
	db := cartTestDB(t)
	svc := NewCartService(db)

	organizer := seedOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedEvent(t, db, organizer.ID)
		defer cleanupCarts(t, db, event.ID)
		defer db.Unscoped().Delete(&event)

		// Insert an expired cart directly (expires_at in the past).
		hoursAgo := rapid.IntRange(25, 72).Draw(rt, "hoursAgo")
		expiredAt := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)

		useUser := rapid.Bool().Draw(rt, "useUser")
		var userID *uuid.UUID
		guestSession := ""
		if useUser {
			uid := uuid.New()
			userID = &uid
		} else {
			guestSession = "guest-" + uuid.New().String()
		}

		cart := models.DraftCart{
			ID:           uuid.New(),
			UserID:       userID,
			GuestSession: guestSession,
			EventID:      event.ID,
			Status:       models.DraftCartActive,
			ExpiresAt:    expiredAt,
		}
		if err := db.Create(&cart).Error; err != nil {
			rt.Fatalf("failed to insert expired cart: %v", err)
		}

		got, err := svc.GetCart(userID, guestSession, event.ID)
		if err != nil {
			rt.Fatalf("GetCart failed: %v", err)
		}
		if got != nil {
			rt.Fatalf("expected nil for expired cart (expires_at=%v), got cart id=%s", expiredAt, got.ID)
		}
	})
}

// ── Property 8 ──────────────────────────────────────────────────────────────

// TestPropertyInvalidCartItemsFilteredOnRestore verifies that
// ValidateAndCleanCart removes items whose ticket type is unavailable
// (sold-out or outside sale window).
//
// Feature: guest-checkout-cart-sharing-metrics, Property 8: Invalid cart items are filtered on restore
//
// Validates: Requirements 2.5
func TestPropertyInvalidCartItemsFilteredOnRestore(t *testing.T) {
	db := cartTestDB(t)
	svc := NewCartService(db)

	organizer := seedOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedEvent(t, db, organizer.ID)
		defer cleanupCarts(t, db, event.ID)
		defer db.Unscoped().Delete(&event)

		now := time.Now()

		// Create one valid ticket type.
		validTT := models.TicketType{
			ID:        uuid.New(),
			EventID:   event.ID,
			Name:      "Valid " + uuid.New().String()[:8],
			Price:     500,
			Quantity:  50,
			Sold:      0,
			SaleStart: now.Add(-time.Hour),
			SaleEnd:   now.Add(24 * time.Hour),
			IsActive:  true,
		}
		if err := db.Create(&validTT).Error; err != nil {
			rt.Fatalf("create validTT: %v", err)
		}
		defer db.Unscoped().Delete(&validTT)

		// Choose an invalid scenario: sold-out or expired sale window.
		invalidKind := rapid.IntRange(0, 1).Draw(rt, "invalidKind")
		invalidTT := models.TicketType{
			ID:       uuid.New(),
			EventID:  event.ID,
			Name:     "Invalid " + uuid.New().String()[:8],
			Price:    500,
			Quantity: 10,
			IsActive: true,
		}
		switch invalidKind {
		case 0: // sold-out
			invalidTT.Sold = 10
			invalidTT.SaleStart = now.Add(-time.Hour)
			invalidTT.SaleEnd = now.Add(24 * time.Hour)
		case 1: // expired sale window
			invalidTT.Sold = 0
			invalidTT.SaleStart = now.Add(-48 * time.Hour)
			invalidTT.SaleEnd = now.Add(-time.Hour)
		}
		if err := db.Create(&invalidTT).Error; err != nil {
			rt.Fatalf("create invalidTT: %v", err)
		}
		defer db.Unscoped().Delete(&invalidTT)

		// Build a cart with both items.
		cartID := uuid.New()
		cart := models.DraftCart{
			ID:        cartID,
			EventID:   event.ID,
			Status:    models.DraftCartActive,
			ExpiresAt: now.Add(24 * time.Hour),
		}
		if err := db.Create(&cart).Error; err != nil {
			rt.Fatalf("create cart: %v", err)
		}

		validItem := models.DraftCartItem{
			ID:           uuid.New(),
			CartID:       cartID,
			TicketTypeID: validTT.ID,
			Quantity:     1,
		}
		invalidItem := models.DraftCartItem{
			ID:           uuid.New(),
			CartID:       cartID,
			TicketTypeID: invalidTT.ID,
			Quantity:     1,
		}
		if err := db.Create(&validItem).Error; err != nil {
			rt.Fatalf("create validItem: %v", err)
		}
		if err := db.Create(&invalidItem).Error; err != nil {
			rt.Fatalf("create invalidItem: %v", err)
		}

		// Load cart with associations (as GetCart would).
		var loadedCart models.DraftCart
		if err := db.Preload("Items.TicketType").First(&loadedCart, "id = ?", cartID).Error; err != nil {
			rt.Fatalf("load cart: %v", err)
		}

		removed, err := svc.ValidateAndCleanCart(&loadedCart)
		if err != nil {
			rt.Fatalf("ValidateAndCleanCart: %v", err)
		}

		// The invalid item's name must appear in removed.
		foundInRemoved := false
		for _, name := range removed {
			if name == invalidTT.Name {
				foundInRemoved = true
				break
			}
		}
		if !foundInRemoved {
			rt.Fatalf("expected %q in removed list, got %v", invalidTT.Name, removed)
		}

		// The invalid ticket type must not appear in the remaining items.
		for _, item := range loadedCart.Items {
			if item.TicketTypeID == invalidTT.ID {
				rt.Fatalf("invalid ticket type %s still present after ValidateAndCleanCart", invalidTT.ID)
			}
		}

		// The valid item must still be present.
		foundValid := false
		for _, item := range loadedCart.Items {
			if item.TicketTypeID == validTT.ID {
				foundValid = true
				break
			}
		}
		if !foundValid {
			rt.Fatal("valid ticket type was incorrectly removed by ValidateAndCleanCart")
		}
	})
}

// ── Property 9 ──────────────────────────────────────────────────────────────

// TestPropertyCartDeletedAfterCheckoutOrClear verifies that after DeleteCart,
// GetCart returns nil for that user-event pair.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 9: Cart deleted after checkout or clear
//
// Validates: Requirements 2.6, 2.7
func TestPropertyCartDeletedAfterCheckoutOrClear(t *testing.T) {
	db := cartTestDB(t)
	svc := NewCartService(db)

	organizer := seedOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedEvent(t, db, organizer.ID)
		defer cleanupCarts(t, db, event.ID)
		defer db.Unscoped().Delete(&event)

		tt := seedTicketType(t, db, event.ID)
		defer db.Unscoped().Delete(&tt)

		useUser := rapid.Bool().Draw(rt, "useUser")
		var userID *uuid.UUID
		guestSession := ""
		if useUser {
			uid := uuid.New()
			userID = &uid
		} else {
			guestSession = "guest-" + uuid.New().String()
		}

		items := []CartItemInput{{TicketTypeID: tt.ID, Quantity: 1}}

		// Create a cart.
		_, err := svc.UpsertCart(userID, guestSession, event.ID, items)
		if err != nil {
			rt.Fatalf("UpsertCart: %v", err)
		}

		// Delete it.
		if err := svc.DeleteCart(userID, guestSession, event.ID); err != nil {
			rt.Fatalf("DeleteCart: %v", err)
		}

		// GetCart must return nil.
		got, err := svc.GetCart(userID, guestSession, event.ID)
		if err != nil {
			rt.Fatalf("GetCart after delete: %v", err)
		}
		if got != nil {
			rt.Fatalf("expected nil after DeleteCart, got cart id=%s status=%s", got.ID, got.Status)
		}
	})
}

// ── Property 10 ─────────────────────────────────────────────────────────────

// TestPropertyAtMostOneActiveCartPerUserEvent verifies that calling UpsertCart
// twice for the same user-event pair results in exactly one active row in the DB.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 10: At most one active cart per user-event pair
//
// Validates: Requirements 2.8
func TestPropertyAtMostOneActiveCartPerUserEvent(t *testing.T) {
	db := cartTestDB(t)
	svc := NewCartService(db)

	organizer := seedOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedEvent(t, db, organizer.ID)
		defer cleanupCarts(t, db, event.ID)
		defer db.Unscoped().Delete(&event)

		tt1 := seedTicketType(t, db, event.ID)
		tt2 := seedTicketType(t, db, event.ID)
		defer db.Unscoped().Delete(&tt1)
		defer db.Unscoped().Delete(&tt2)

		useUser := rapid.Bool().Draw(rt, "useUser")
		var userID *uuid.UUID
		guestSession := ""
		if useUser {
			uid := uuid.New()
			userID = &uid
		} else {
			guestSession = "guest-" + uuid.New().String()
		}

		items1 := []CartItemInput{{TicketTypeID: tt1.ID, Quantity: 1}}
		items2 := []CartItemInput{{TicketTypeID: tt2.ID, Quantity: 2}}

		// First upsert.
		if _, err := svc.UpsertCart(userID, guestSession, event.ID, items1); err != nil {
			rt.Fatalf("first UpsertCart: %v", err)
		}

		// Second upsert (same user+event, different items).
		if _, err := svc.UpsertCart(userID, guestSession, event.ID, items2); err != nil {
			rt.Fatalf("second UpsertCart: %v", err)
		}

		// Count active carts for this user-event pair.
		var count int64
		query := db.Model(&models.DraftCart{}).
			Where("event_id = ? AND status = ?", event.ID, models.DraftCartActive)
		if userID != nil {
			query = query.Where("user_id = ?", userID)
		} else {
			query = query.Where("guest_session = ?", guestSession)
		}
		if err := query.Count(&count).Error; err != nil {
			rt.Fatalf("count active carts: %v", err)
		}

		if count != 1 {
			rt.Fatalf("expected exactly 1 active cart, found %d", count)
		}
	})
}
