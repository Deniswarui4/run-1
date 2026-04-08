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

// metricsTestDB opens a GORM connection to the test database and auto-migrates
// all tables needed by MetricsService tests.
func metricsTestDB(t *testing.T) *gorm.DB {
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

// seedMetricsOrganizer inserts a minimal organizer User.
func seedMetricsOrganizer(t *testing.T, db *gorm.DB) models.User {
	t.Helper()
	u := models.User{
		ID:        uuid.New(),
		Email:     "metrics-org-" + uuid.New().String() + "@test.com",
		Password:  "hashed",
		FirstName: "Metrics",
		LastName:  "Organizer",
		Role:      models.RoleOrganizer,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seedMetricsOrganizer: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&u) })
	return u
}

// seedMetricsEvent inserts a published Event.
func seedMetricsEvent(t *testing.T, db *gorm.DB, organizerID uuid.UUID) models.Event {
	t.Helper()
	now := time.Now()
	e := models.Event{
		ID:          uuid.New(),
		Title:       "Metrics Event " + uuid.New().String()[:8],
		Venue:       "Test Venue",
		StartDate:   now.Add(24 * time.Hour),
		EndDate:     now.Add(48 * time.Hour),
		Status:      models.EventStatusPublished,
		OrganizerID: organizerID,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seedMetricsEvent: %v", err)
	}
	t.Cleanup(func() {
		db.Unscoped().Where("event_id = ?", e.ID).Delete(&models.Ticket{})
		db.Unscoped().Where("event_id = ?", e.ID).Delete(&models.TicketType{})
		db.Unscoped().Delete(&e)
	})
	return e
}

// seedMetricsTicketType inserts a TicketType with the given quantity and price.
func seedMetricsTicketType(t *testing.T, db *gorm.DB, eventID uuid.UUID, name string, quantity int, price float64) models.TicketType {
	t.Helper()
	now := time.Now()
	tt := models.TicketType{
		ID:          uuid.New(),
		EventID:     eventID,
		Name:        name,
		Price:       price,
		Quantity:    quantity,
		Sold:        0,
		MaxPerOrder: 10,
		SaleStart:   now.Add(-time.Hour),
		SaleEnd:     now.Add(24 * time.Hour),
		IsActive:    true,
	}
	if err := db.Create(&tt).Error; err != nil {
		t.Fatalf("seedMetricsTicketType: %v", err)
	}
	return tt
}

// seedTransaction inserts a completed Transaction for the given event and user.
func seedTransaction(t *testing.T, db *gorm.DB, userID, eventID uuid.UUID, amount float64) models.Transaction {
	t.Helper()
	tx := models.Transaction{
		ID:               uuid.New(),
		UserID:           userID,
		EventID:          &eventID,
		Type:             models.TransactionTypeTicketPurchase,
		Status:           models.TransactionStatusCompleted,
		Amount:           amount,
		NetAmount:        amount,
		PaymentGateway:   "paystack",
		PaymentReference: "ref-" + uuid.New().String(),
	}
	if err := db.Create(&tx).Error; err != nil {
		t.Fatalf("seedTransaction: %v", err)
	}
	return tx
}

// seedConfirmedTicket inserts a confirmed Ticket for the given event, ticket type, and transaction.
// createdAt overrides the ticket's creation timestamp (used for daily revenue tests).
func seedConfirmedTicket(t *testing.T, db *gorm.DB, eventID, ticketTypeID, transactionID uuid.UUID, price float64, createdAt time.Time, checkedIn bool) models.Ticket {
	t.Helper()
	var checkedInAt *time.Time
	if checkedIn {
		ts := createdAt.Add(time.Hour)
		checkedInAt = &ts
	}
	tk := models.Ticket{
		ID:            uuid.New(),
		TicketNumber:  "TKT-" + uuid.New().String()[:8],
		EventID:       eventID,
		TicketTypeID:  ticketTypeID,
		TransactionID: transactionID,
		Status:        models.TicketStatusConfirmed,
		Price:         price,
		CheckedInAt:   checkedInAt,
		CreatedAt:     createdAt,
	}
	if err := db.Create(&tk).Error; err != nil {
		t.Fatalf("seedConfirmedTicket: %v", err)
	}
	return tk
}

// ── Property 15 ─────────────────────────────────────────────────────────────

// TestPropertyTicketTypeBreakdownCoversAllTypes verifies that
// GetEventTicketTypeBreakdown returns exactly N entries for an event with N
// ticket types, with correct sold/remaining/revenue values.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 15: Ticket type breakdown covers all ticket types
//
// Validates: Requirements 4.1
func TestPropertyTicketTypeBreakdownCoversAllTypes(t *testing.T) {
	db := metricsTestDB(t)
	svc := NewMetricsService(db)
	organizer := seedMetricsOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedMetricsEvent(t, db, organizer.ID)

		// Generate N ticket types (1–5).
		n := rapid.IntRange(1, 5).Draw(rt, "numTypes")

		type ttInfo struct {
			tt       models.TicketType
			sold     int
			price    float64
			quantity int
		}
		infos := make([]ttInfo, n)

		for i := 0; i < n; i++ {
			qty := rapid.IntRange(10, 100).Draw(rt, fmt.Sprintf("qty%d", i))
			price := float64(rapid.IntRange(100, 5000).Draw(rt, fmt.Sprintf("price%d", i)))
			name := fmt.Sprintf("Type-%d-%s", i, uuid.New().String()[:6])
			tt := seedMetricsTicketType(t, db, event.ID, name, qty, price)
			infos[i] = ttInfo{tt: tt, price: price, quantity: qty}
		}

		// Seed a system user for transactions.
		sysUser := models.User{
			ID: uuid.New(), Email: "sys-" + uuid.New().String() + "@test.com",
			Password: "x", FirstName: "S", LastName: "U", Role: models.RoleAttendee,
		}
		db.Create(&sysUser)
		defer db.Unscoped().Delete(&sysUser)

		// Seed confirmed tickets for each type.
		now := time.Now()
		for i := range infos {
			soldCount := rapid.IntRange(0, infos[i].quantity).Draw(rt, fmt.Sprintf("sold%d", i))
			infos[i].sold = soldCount
			if soldCount == 0 {
				continue
			}
			tx := seedTransaction(t, db, sysUser.ID, event.ID, float64(soldCount)*infos[i].price)
			defer db.Unscoped().Delete(&tx)
			for j := 0; j < soldCount; j++ {
				tk := seedConfirmedTicket(t, db, event.ID, infos[i].tt.ID, tx.ID, infos[i].price, now, false)
				defer db.Unscoped().Delete(&tk)
			}
		}

		breakdown, err := svc.GetEventTicketTypeBreakdown(event.ID)
		if err != nil {
			rt.Fatalf("GetEventTicketTypeBreakdown: %v", err)
		}

		// Must have exactly N entries.
		if len(breakdown) != n {
			rt.Fatalf("expected %d entries, got %d", n, len(breakdown))
		}

		// Build lookup by ticket type ID.
		byID := make(map[uuid.UUID]TicketTypeBreakdown, len(breakdown))
		for _, b := range breakdown {
			byID[b.TicketTypeID] = b
		}

		for _, info := range infos {
			b, ok := byID[info.tt.ID]
			if !ok {
				rt.Fatalf("missing breakdown entry for ticket type %s", info.tt.ID)
			}
			if b.Sold != int64(info.sold) {
				rt.Fatalf("type %s: sold want %d got %d", info.tt.ID, info.sold, b.Sold)
			}
			expectedRemaining := int64(info.quantity - info.sold)
			if b.Remaining != expectedRemaining {
				rt.Fatalf("type %s: remaining want %d got %d", info.tt.ID, expectedRemaining, b.Remaining)
			}
			expectedRevenue := float64(info.sold) * info.price
			if b.GrossRevenue != expectedRevenue {
				rt.Fatalf("type %s: revenue want %.2f got %.2f", info.tt.ID, expectedRevenue, b.GrossRevenue)
			}
		}
	})
}

// ── Property 16 ─────────────────────────────────────────────────────────────

// TestPropertyDailyRevenueTimeSeriesHasNoGaps verifies that
// GetEventDailyRevenue returns one entry per calendar day from the earliest
// sale date to today, with zero revenue on days with no sales.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 16: Daily revenue time series has no gaps
//
// Validates: Requirements 4.2
func TestPropertyDailyRevenueTimeSeriesHasNoGaps(t *testing.T) {
	db := metricsTestDB(t)
	svc := NewMetricsService(db)
	organizer := seedMetricsOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedMetricsEvent(t, db, organizer.ID)
		tt := seedMetricsTicketType(t, db, event.ID, "GA-"+uuid.New().String()[:6], 1000, 500)
		defer db.Unscoped().Delete(&tt)

		sysUser := models.User{
			ID: uuid.New(), Email: "sys2-" + uuid.New().String() + "@test.com",
			Password: "x", FirstName: "S", LastName: "U", Role: models.RoleAttendee,
		}
		db.Create(&sysUser)
		defer db.Unscoped().Delete(&sysUser)

		// Seed tickets on a random subset of days within the last 1–14 days.
		spanDays := rapid.IntRange(1, 14).Draw(rt, "spanDays")
		now := time.Now().Truncate(24 * time.Hour)
		startDay := now.AddDate(0, 0, -spanDays)

		// Pick which days have sales (at least the first day must have a sale
		// so the series has a defined start).
		saleDays := make(map[int]bool)
		saleDays[0] = true // always seed the first day
		for d := 1; d <= spanDays; d++ {
			if rapid.Bool().Draw(rt, fmt.Sprintf("saleDay%d", d)) {
				saleDays[d] = true
			}
		}

		tx := seedTransaction(t, db, sysUser.ID, event.ID, 0)
		defer db.Unscoped().Delete(&tx)

		var seededTickets []models.Ticket
		for d := range saleDays {
			saleTime := startDay.AddDate(0, 0, d).Add(time.Hour) // noon-ish
			tk := seedConfirmedTicket(t, db, event.ID, tt.ID, tx.ID, 500, saleTime, false)
			seededTickets = append(seededTickets, tk)
		}
		defer func() {
			for _, tk := range seededTickets {
				db.Unscoped().Delete(&tk)
			}
		}()

		series, err := svc.GetEventDailyRevenue(event.ID)
		if err != nil {
			rt.Fatalf("GetEventDailyRevenue: %v", err)
		}

		if len(series) == 0 {
			rt.Fatal("expected non-empty series")
		}

		// Parse first and last dates.
		firstDate, _ := time.Parse("2006-01-02", series[0].Date)
		lastDate, _ := time.Parse("2006-01-02", series[len(series)-1].Date)

		// Verify no gaps: consecutive dates.
		for i := 1; i < len(series); i++ {
			prev, _ := time.Parse("2006-01-02", series[i-1].Date)
			curr, _ := time.Parse("2006-01-02", series[i].Date)
			if !curr.Equal(prev.AddDate(0, 0, 1)) {
				rt.Fatalf("gap between %s and %s", series[i-1].Date, series[i].Date)
			}
		}

		// Last entry must be today or earlier.
		today := time.Now().Truncate(24 * time.Hour)
		if lastDate.After(today) {
			rt.Fatalf("last date %s is after today %s", lastDate.Format("2006-01-02"), today.Format("2006-01-02"))
		}

		// Build a set of dates that should have revenue.
		revenueByDate := make(map[string]float64, len(series))
		for _, entry := range series {
			revenueByDate[entry.Date] = entry.Revenue
		}

		// Days with sales must have revenue > 0; days without must be 0.
		for d := 0; d <= spanDays; d++ {
			day := startDay.AddDate(0, 0, d)
			if day.After(today) {
				break
			}
			key := day.Format("2006-01-02")
			// Only check days within the series range.
			if day.Before(firstDate) || day.After(lastDate) {
				continue
			}
			rev, exists := revenueByDate[key]
			if !exists {
				rt.Fatalf("missing entry for date %s", key)
			}
			if saleDays[d] && rev == 0 {
				rt.Fatalf("expected revenue > 0 on sale day %s, got 0", key)
			}
			if !saleDays[d] && rev != 0 {
				rt.Fatalf("expected revenue = 0 on non-sale day %s, got %.2f", key, rev)
			}
		}
	})
}

// ── Property 17 ─────────────────────────────────────────────────────────────

// TestPropertyCheckInRateInvariant verifies that GetEventCheckInStats returns
// check_in_rate = checked_in / total_confirmed and checked_in <= total_confirmed.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 17: Check-in rate invariant
//
// Validates: Requirements 4.3
func TestPropertyCheckInRateInvariant(t *testing.T) {
	db := metricsTestDB(t)
	svc := NewMetricsService(db)
	organizer := seedMetricsOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		event := seedMetricsEvent(t, db, organizer.ID)
		tt := seedMetricsTicketType(t, db, event.ID, "CI-"+uuid.New().String()[:6], 500, 200)
		defer db.Unscoped().Delete(&tt)

		sysUser := models.User{
			ID: uuid.New(), Email: "sys3-" + uuid.New().String() + "@test.com",
			Password: "x", FirstName: "S", LastName: "U", Role: models.RoleAttendee,
		}
		db.Create(&sysUser)
		defer db.Unscoped().Delete(&sysUser)

		totalConfirmed := rapid.IntRange(0, 50).Draw(rt, "totalConfirmed")
		checkedIn := 0
		if totalConfirmed > 0 {
			checkedIn = rapid.IntRange(0, totalConfirmed).Draw(rt, "checkedIn")
		}

		now := time.Now()
		tx := seedTransaction(t, db, sysUser.ID, event.ID, float64(totalConfirmed)*200)
		defer db.Unscoped().Delete(&tx)

		var tickets []models.Ticket
		for i := 0; i < totalConfirmed; i++ {
			isCheckedIn := i < checkedIn
			tk := seedConfirmedTicket(t, db, event.ID, tt.ID, tx.ID, 200, now, isCheckedIn)
			tickets = append(tickets, tk)
		}
		defer func() {
			for _, tk := range tickets {
				db.Unscoped().Delete(&tk)
			}
		}()

		stats, err := svc.GetEventCheckInStats(event.ID)
		if err != nil {
			rt.Fatalf("GetEventCheckInStats: %v", err)
		}

		if stats.TotalConfirmed != int64(totalConfirmed) {
			rt.Fatalf("TotalConfirmed: want %d got %d", totalConfirmed, stats.TotalConfirmed)
		}
		if stats.CheckedIn != int64(checkedIn) {
			rt.Fatalf("CheckedIn: want %d got %d", checkedIn, stats.CheckedIn)
		}
		// checked_in must never exceed total_confirmed.
		if stats.CheckedIn > stats.TotalConfirmed {
			rt.Fatalf("CheckedIn (%d) > TotalConfirmed (%d)", stats.CheckedIn, stats.TotalConfirmed)
		}
		// Rate invariant.
		if totalConfirmed == 0 {
			if stats.CheckInRate != 0.0 {
				rt.Fatalf("expected rate 0.0 when no confirmed tickets, got %f", stats.CheckInRate)
			}
		} else {
			expectedRate := float64(checkedIn) / float64(totalConfirmed)
			const epsilon = 1e-9
			diff := stats.CheckInRate - expectedRate
			if diff < -epsilon || diff > epsilon {
				rt.Fatalf("CheckInRate: want %f got %f", expectedRate, stats.CheckInRate)
			}
		}
	})
}

// ── Property 18 ─────────────────────────────────────────────────────────────

// TestPropertyTopEventsRankingOrderedAndBounded verifies that
// GetTopEventsByRevenue returns at most 10 results ordered by gross revenue
// descending, and every event in the list has revenue >= any event not in it.
//
// Feature: guest-checkout-cart-sharing-metrics, Property 18: Top events ranking is correctly ordered and bounded
//
// Validates: Requirements 4.6
func TestPropertyTopEventsRankingOrderedAndBounded(t *testing.T) {
	db := metricsTestDB(t)
	svc := NewMetricsService(db)
	organizer := seedMetricsOrganizer(t, db)

	rapid.Check(t, func(rt *rapid.T) {
		// Seed 1–15 events with known revenues.
		numEvents := rapid.IntRange(1, 15).Draw(rt, "numEvents")

		type eventInfo struct {
			event   models.Event
			revenue float64
		}
		infos := make([]eventInfo, numEvents)

		sysUser := models.User{
			ID: uuid.New(), Email: "sys4-" + uuid.New().String() + "@test.com",
			Password: "x", FirstName: "S", LastName: "U", Role: models.RoleAttendee,
		}
		db.Create(&sysUser)
		defer db.Unscoped().Delete(&sysUser)

		now := time.Now()
		start := now.AddDate(0, 0, -7)
		end := now.AddDate(0, 0, 1)

		for i := 0; i < numEvents; i++ {
			e := seedMetricsEvent(t, db, organizer.ID)
			tt := seedMetricsTicketType(t, db, e.ID, "T-"+uuid.New().String()[:6], 1000, 100)
			defer db.Unscoped().Delete(&tt)

			revenue := float64(rapid.IntRange(0, 10000).Draw(rt, fmt.Sprintf("rev%d", i)))
			infos[i] = eventInfo{event: e, revenue: revenue}

			if revenue > 0 {
				ticketCount := int(revenue / 100)
				tx := seedTransaction(t, db, sysUser.ID, e.ID, revenue)
				defer db.Unscoped().Delete(&tx)
				for j := 0; j < ticketCount; j++ {
					tk := seedConfirmedTicket(t, db, e.ID, tt.ID, tx.ID, 100, now, false)
					defer db.Unscoped().Delete(&tk)
				}
			}
		}

		results, err := svc.GetTopEventsByRevenue(start, end, 10)
		if err != nil {
			rt.Fatalf("GetTopEventsByRevenue: %v", err)
		}

		// Must return at most 10 results.
		if len(results) > 10 {
			rt.Fatalf("expected at most 10 results, got %d", len(results))
		}

		// Must be ordered by gross revenue descending.
		for i := 1; i < len(results); i++ {
			if results[i].GrossRevenue > results[i-1].GrossRevenue {
				rt.Fatalf("results not ordered: index %d (%.2f) > index %d (%.2f)",
					i, results[i].GrossRevenue, i-1, results[i-1].GrossRevenue)
			}
		}

		// Every event in the result must have revenue >= any event not in the result.
		if len(results) > 0 && numEvents > len(results) {
			minInList := results[len(results)-1].GrossRevenue

			// Build set of returned event IDs.
			inList := make(map[uuid.UUID]bool, len(results))
			for _, r := range results {
				inList[r.EventID] = true
			}

			// Check events seeded by this iteration that are NOT in the list.
			for _, info := range infos {
				if inList[info.event.ID] {
					continue
				}
				if info.revenue > minInList {
					rt.Fatalf(
						"event %s with revenue %.2f not in top list but min in list is %.2f",
						info.event.ID, info.revenue, minInList,
					)
				}
			}
		}
	})
}
