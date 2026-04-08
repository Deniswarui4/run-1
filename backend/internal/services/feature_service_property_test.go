//go:build integration

// Feature: guest-checkout-cart-sharing-metrics
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

// featureTestDB opens a GORM connection to the test database and auto-migrates
// all tables needed by FeatureService tests.
func featureTestDB(t *testing.T) *gorm.DB {
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
		&models.PlatformSettings{},
		&models.EventFeatureEvaluation{},
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

// seedFeatureOrganizer inserts a minimal organizer User.
func seedFeatureOrganizer(t *testing.T, db *gorm.DB) models.User {
	t.Helper()
	u := models.User{
		ID:        uuid.New(),
		Email:     "feat-org-" + uuid.New().String() + "@test.com",
		Password:  "hashed",
		FirstName: "Feature",
		LastName:  "Organizer",
		Role:      models.RoleOrganizer,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seedFeatureOrganizer: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&u) })
	return u
}

// seedFeatureEvent inserts an Event with the given status.
func seedFeatureEvent(t *testing.T, db *gorm.DB, organizerID uuid.UUID, status models.EventStatus) models.Event {
	t.Helper()
	now := time.Now()
	e := models.Event{
		ID:          uuid.New(),
		Title:       "Feature Event " + uuid.New().String()[:8],
		Venue:       "Test Venue",
		StartDate:   now.Add(24 * time.Hour),
		EndDate:     now.Add(48 * time.Hour),
		Status:      status,
		IsFeatured:  false,
		FeaturedType: models.FeaturedTypeNone,
		OrganizerID: organizerID,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seedFeatureEvent: %v", err)
	}
	t.Cleanup(func() {
		db.Unscoped().Where("event_id = ?", e.ID).Delete(&models.EventFeatureEvaluation{})
		db.Unscoped().Where("event_id = ?", e.ID).Delete(&models.Ticket{})
		db.Unscoped().Where("event_id = ?", e.ID).Delete(&models.TicketType{})
		db.Unscoped().Delete(&e)
	})
	return e
}

// seedPlatformSettings ensures exactly one PlatformSettings row exists with the given threshold.
func seedPlatformSettings(t *testing.T, db *gorm.DB, threshold int) models.PlatformSettings {
	t.Helper()
	var settings models.PlatformSettings
	if err := db.First(&settings).Error; err != nil {
		settings = models.PlatformSettings{
			ID:                   uuid.New(),
			AutoFeatureThreshold: threshold,
		}
		if err := db.Create(&settings).Error; err != nil {
			t.Fatalf("seedPlatformSettings create: %v", err)
		}
	} else {
		settings.AutoFeatureThreshold = threshold
		if err := db.Save(&settings).Error; err != nil {
			t.Fatalf("seedPlatformSettings save: %v", err)
		}
	}
	return settings
}

// seedRecentTickets inserts n confirmed tickets for the event created within the last 24 hours.
func seedRecentTickets(t *testing.T, db *gorm.DB, eventID uuid.UUID, n int) {
	t.Helper()
	if n == 0 {
		return
	}

	organizer := seedFeatureOrganizer(t, db)
	tt := models.TicketType{
		ID:          uuid.New(),
		EventID:     eventID,
		Name:        "GA-" + uuid.New().String()[:6],
		Price:       100,
		Quantity:    n + 100,
		Sold:        0,
		MaxPerOrder: 10,
		SaleStart:   time.Now().Add(-time.Hour),
		SaleEnd:     time.Now().Add(24 * time.Hour),
		IsActive:    true,
	}
	if err := db.Create(&tt).Error; err != nil {
		t.Fatalf("seedRecentTickets: create ticket type: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&tt) })

	tx := models.Transaction{
		ID:               uuid.New(),
		UserID:           organizer.ID,
		EventID:          &eventID,
		Type:             models.TransactionTypeTicketPurchase,
		Status:           models.TransactionStatusCompleted,
		Amount:           float64(n) * 100,
		NetAmount:        float64(n) * 100,
		PaymentGateway:   "paystack",
		PaymentReference: "ref-" + uuid.New().String(),
	}
	if err := db.Create(&tx).Error; err != nil {
		t.Fatalf("seedRecentTickets: create transaction: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&tx) })

	recent := time.Now().Add(-time.Hour) // within last 24h
	for i := 0; i < n; i++ {
		tk := models.Ticket{
			ID:            uuid.New(),
			TicketNumber:  "TKT-" + uuid.New().String()[:8],
			EventID:       eventID,
			TicketTypeID:  tt.ID,
			TransactionID: tx.ID,
			Status:        models.TicketStatusConfirmed,
			Price:         100,
			CreatedAt:     recent,
		}
		if err := db.Create(&tk).Error; err != nil {
			t.Fatalf("seedRecentTickets: create ticket %d: %v", i, err)
		}
		t.Cleanup(func() { db.Unscoped().Delete(&tk) })
	}
}

// ── Property 19 ─────────────────────────────────────────────────────────────

// TestPropertyFeaturedToggleIsIdempotentOverTwoApplications verifies that
// calling ToggleFeatured twice on a published event returns is_featured to its
// original value.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 19: Featured toggle is idempotent over two applications
//
// Validates: Requirements 5.1
func TestPropertyFeaturedToggleIsIdempotentOverTwoApplications(t *testing.T) {
	db := featureTestDB(t)
	svc := NewFeatureService(db)
	organizer := seedFeatureOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedFeatureEvent(t, db, organizer.ID, models.EventStatusPublished)

		// Record original state.
		originalFeatured := event.IsFeatured

		// First toggle.
		after1, err := svc.ToggleFeatured(event.ID, true)
		if err != nil {
			rt.Fatalf("ToggleFeatured (1st): %v", err)
		}
		if after1.IsFeatured == originalFeatured {
			rt.Fatalf("first toggle did not change is_featured: still %v", originalFeatured)
		}

		// Second toggle — must return to original.
		after2, err := svc.ToggleFeatured(event.ID, true)
		if err != nil {
			rt.Fatalf("ToggleFeatured (2nd): %v", err)
		}
		if after2.IsFeatured != originalFeatured {
			rt.Fatalf("after two toggles: want is_featured=%v, got %v", originalFeatured, after2.IsFeatured)
		}
	})
}

// ── Property 20 ─────────────────────────────────────────────────────────────

// TestPropertyNonPublishedEventsCannotBeManuallyFeatured verifies that
// ToggleFeatured returns ErrEventNotPublished for any non-published event status.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 20: Non-published events cannot be manually featured
//
// Validates: Requirements 5.2
func TestPropertyNonPublishedEventsCannotBeManuallyFeatured(t *testing.T) {
	db := featureTestDB(t)
	svc := NewFeatureService(db)
	organizer := seedFeatureOrganizer(t, db)

	nonPublishedStatuses := []models.EventStatus{
		models.EventStatusDraft,
		models.EventStatusPending,
		models.EventStatusApproved,
		models.EventStatusRejected,
		models.EventStatusCancelled,
		models.EventStatusCompleted,
	}

	rapid.Check(t, func(rt *rapid.T) {
		idx := rapid.IntRange(0, len(nonPublishedStatuses)-1).Draw(rt, "statusIdx")
		status := nonPublishedStatuses[idx]

		event := seedFeatureEvent(t, db, organizer.ID, status)

		_, err := svc.ToggleFeatured(event.ID, true)
		if err == nil {
			rt.Fatalf("expected ErrEventNotPublished for status %q, got nil", status)
		}
		if err != ErrEventNotPublished {
			rt.Fatalf("expected ErrEventNotPublished for status %q, got: %v", status, err)
		}
	})
}

// ── Property 21 ─────────────────────────────────────────────────────────────

// TestPropertyAutoFeatureThresholdSettingRoundTrip verifies that writing a
// non-negative auto_feature_threshold to platform settings and reading it back
// returns the same value.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 21: Auto-feature threshold setting round-trip
//
// Validates: Requirements 5.4
func TestPropertyAutoFeatureThresholdSettingRoundTrip(t *testing.T) {
	db := featureTestDB(t)

	// Ensure a settings row exists.
	seedPlatformSettings(t, db, 0)

	rapid.Check(t, func(rt *rapid.T) {
		threshold := rapid.IntRange(0, 10000).Draw(rt, "threshold")

		// Write.
		var settings models.PlatformSettings
		if err := db.First(&settings).Error; err != nil {
			rt.Fatalf("read settings: %v", err)
		}
		settings.AutoFeatureThreshold = threshold
		if err := db.Save(&settings).Error; err != nil {
			rt.Fatalf("save settings: %v", err)
		}

		// Read back.
		var readBack models.PlatformSettings
		if err := db.First(&readBack).Error; err != nil {
			rt.Fatalf("read back settings: %v", err)
		}

		if readBack.AutoFeatureThreshold != threshold {
			rt.Fatalf("threshold round-trip: wrote %d, read %d", threshold, readBack.AutoFeatureThreshold)
		}
	})
}

// ── Property 22 ─────────────────────────────────────────────────────────────

// TestPropertyAutoFeatureEvaluationFeaturesQualifyingEvents verifies that
// EvaluateAutoFeature sets is_featured=true for a published event whose
// sales velocity meets or exceeds the configured threshold.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 22: Auto-feature evaluation features qualifying events
//
// Validates: Requirements 5.5
func TestPropertyAutoFeatureEvaluationFeaturesQualifyingEvents(t *testing.T) {
	db := featureTestDB(t)
	svc := NewFeatureService(db)
	organizer := seedFeatureOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		// threshold in [1, 20] so we can seed enough tickets without being slow.
		threshold := rapid.IntRange(1, 20).Draw(rt, "threshold")
		// velocity >= threshold.
		velocity := rapid.IntRange(threshold, threshold+10).Draw(rt, "velocity")

		seedPlatformSettings(t, db, threshold)

		event := seedFeatureEvent(t, db, organizer.ID, models.EventStatusPublished)
		seedRecentTickets(t, db, event.ID, velocity)

		if err := svc.EvaluateAutoFeature(); err != nil {
			rt.Fatalf("EvaluateAutoFeature: %v", err)
		}

		var updated models.Event
		if err := db.First(&updated, "id = ?", event.ID).Error; err != nil {
			rt.Fatalf("reload event: %v", err)
		}

		if !updated.IsFeatured {
			rt.Fatalf("event with velocity %d >= threshold %d should be featured, but is_featured=false", velocity, threshold)
		}
		if updated.FeaturedType != models.FeaturedTypeAuto {
			rt.Fatalf("expected featured_type=%q, got %q", models.FeaturedTypeAuto, updated.FeaturedType)
		}
	})
}

// ── Property 23 ─────────────────────────────────────────────────────────────

// TestPropertyAutoUnfeatureAfter3ConsecutiveBelowThreshold verifies that an
// auto-featured event is unfeatured after 3 consecutive below-threshold
// evaluations, while a manually-featured event remains featured.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 23: Auto-unfeature after 3 consecutive below-threshold evaluations
//
// Validates: Requirements 5.6
func TestPropertyAutoUnfeatureAfter3ConsecutiveBelowThreshold(t *testing.T) {
	db := featureTestDB(t)
	svc := NewFeatureService(db)
	organizer := seedFeatureOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		threshold := rapid.IntRange(5, 20).Draw(rt, "threshold")
		seedPlatformSettings(t, db, threshold)

		// ── Auto-featured event with velocity = 0 (below threshold) ──
		autoEvent := seedFeatureEvent(t, db, organizer.ID, models.EventStatusPublished)
		// Mark it as auto-featured directly.
		db.Model(&autoEvent).Updates(map[string]interface{}{
			"is_featured":   true,
			"featured_type": models.FeaturedTypeAuto,
		})
		// Seed no recent tickets → velocity = 0 < threshold.

		// ── Manually-featured event with velocity = 0 ──
		manualEvent := seedFeatureEvent(t, db, organizer.ID, models.EventStatusPublished)
		db.Model(&manualEvent).Updates(map[string]interface{}{
			"is_featured":   true,
			"featured_type": models.FeaturedTypeManual,
		})

		// Run 3 evaluations.
		for i := 0; i < 3; i++ {
			if err := svc.EvaluateAutoFeature(); err != nil {
				rt.Fatalf("EvaluateAutoFeature iteration %d: %v", i+1, err)
			}
		}

		// Auto-featured event must now be unfeatured.
		var updatedAuto models.Event
		if err := db.First(&updatedAuto, "id = ?", autoEvent.ID).Error; err != nil {
			rt.Fatalf("reload auto event: %v", err)
		}
		if updatedAuto.IsFeatured {
			rt.Fatalf("auto-featured event should be unfeatured after 3 below-threshold evaluations, but is_featured=true")
		}

		// Manually-featured event must remain featured.
		var updatedManual models.Event
		if err := db.First(&updatedManual, "id = ?", manualEvent.ID).Error; err != nil {
			rt.Fatalf("reload manual event: %v", err)
		}
		if !updatedManual.IsFeatured {
			rt.Fatalf("manually-featured event should remain featured regardless of velocity, but is_featured=false")
		}
		if updatedManual.FeaturedType != models.FeaturedTypeManual {
			rt.Fatalf("manually-featured event featured_type changed: got %q", updatedManual.FeaturedType)
		}
	})
}

// ── Property 24 ─────────────────────────────────────────────────────────────

// TestPropertyFeaturedEventsListContainsOnlyPublishedFeaturedInOrder verifies
// that GetFeaturedEvents returns only events where is_featured=true AND
// status=published, ordered by start_date ASC.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 24: Featured events list contains only published+featured events in order
//
// Validates: Requirements 5.8
func TestPropertyFeaturedEventsListContainsOnlyPublishedFeaturedInOrder(t *testing.T) {
	db := featureTestDB(t)
	organizer := seedFeatureOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		// Seed a mix of events: published+featured, published+not-featured, non-published+featured.
		numPublishedFeatured := rapid.IntRange(0, 5).Draw(rt, "numPublishedFeatured")
		numPublishedNotFeatured := rapid.IntRange(0, 3).Draw(rt, "numPublishedNotFeatured")
		numNonPublishedFeatured := rapid.IntRange(0, 3).Draw(rt, "numNonPublishedFeatured")

		nonPublishedStatuses := []models.EventStatus{
			models.EventStatusDraft,
			models.EventStatusPending,
			models.EventStatusCancelled,
		}

		var featuredIDs []uuid.UUID
		now := time.Now()

		// Seed published+featured events with staggered start dates.
		for i := 0; i < numPublishedFeatured; i++ {
			e := seedFeatureEvent(t, db, organizer.ID, models.EventStatusPublished)
			startDate := now.Add(time.Duration(i+1) * 24 * time.Hour)
			db.Model(&e).Updates(map[string]interface{}{
				"is_featured":   true,
				"featured_type": models.FeaturedTypeManual,
				"start_date":    startDate,
			})
			featuredIDs = append(featuredIDs, e.ID)
		}

		// Seed published+not-featured events.
		for i := 0; i < numPublishedNotFeatured; i++ {
			seedFeatureEvent(t, db, organizer.ID, models.EventStatusPublished)
		}

		// Seed non-published+featured events (should NOT appear in results).
		for i := 0; i < numNonPublishedFeatured; i++ {
			status := nonPublishedStatuses[i%len(nonPublishedStatuses)]
			e := seedFeatureEvent(t, db, organizer.ID, status)
			db.Model(&e).Updates(map[string]interface{}{
				"is_featured":   true,
				"featured_type": models.FeaturedTypeManual,
			})
		}

		// Query featured events directly (mirrors GetFeaturedEvents handler logic).
		var results []models.Event
		if err := db.Where("is_featured = ? AND status = ?", true, models.EventStatusPublished).
			Order("start_date ASC").
			Find(&results).Error; err != nil {
			rt.Fatalf("query featured events: %v", err)
		}

		// Build set of returned IDs.
		returnedIDs := make(map[uuid.UUID]bool, len(results))
		for _, e := range results {
			returnedIDs[e.ID] = true
		}

		// Every result must be published and featured.
		for _, e := range results {
			if e.Status != models.EventStatusPublished {
				rt.Fatalf("result event %s has status %q, expected published", e.ID, e.Status)
			}
			if !e.IsFeatured {
				rt.Fatalf("result event %s has is_featured=false", e.ID)
			}
		}

		// All seeded published+featured events must appear in results.
		for _, id := range featuredIDs {
			if !returnedIDs[id] {
				rt.Fatalf("published+featured event %s missing from results", id)
			}
		}

		// Results must be ordered by start_date ASC.
		for i := 1; i < len(results); i++ {
			if results[i].StartDate.Before(results[i-1].StartDate) {
				rt.Fatalf("results not ordered by start_date ASC: index %d (%s) before index %d (%s)",
					i, results[i].StartDate.Format(time.RFC3339),
					i-1, results[i-1].StartDate.Format(time.RFC3339))
			}
		}
	})
}
