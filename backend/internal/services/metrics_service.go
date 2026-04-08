// Feature: guest-checkout-cart-sharing-metrics
package services

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TicketTypeBreakdown holds per-ticket-type aggregated metrics for an event.
type TicketTypeBreakdown struct {
	TicketTypeID uuid.UUID `json:"ticket_type_id"`
	Name         string    `json:"name"`
	Sold         int64     `json:"sold"`
	Remaining    int64     `json:"remaining"`
	GrossRevenue float64   `json:"gross_revenue"`
	CapacityPct  float64   `json:"capacity_pct"` // sold / quantity * 100
}

// DailyRevenue holds revenue for a single calendar day.
type DailyRevenue struct {
	Date    string  `json:"date"` // "YYYY-MM-DD"
	Revenue float64 `json:"revenue"`
}

// CheckInStats holds check-in aggregates for an event.
type CheckInStats struct {
	TotalConfirmed int64   `json:"total_confirmed"`
	CheckedIn      int64   `json:"checked_in"`
	CheckInRate    float64 `json:"check_in_rate"` // checked_in / total_confirmed
}

// DailyCount holds a count for a single calendar day.
type DailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// EventRevenue holds gross revenue summary for a single event.
type EventRevenue struct {
	EventID      uuid.UUID `json:"event_id"`
	Title        string    `json:"title"`
	GrossRevenue float64   `json:"gross_revenue"`
	TicketsSold  int64     `json:"tickets_sold"`
}

// MetricsService computes analytics for organizer and admin dashboards.
type MetricsService struct {
	db *gorm.DB
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(db *gorm.DB) *MetricsService {
	return &MetricsService{db: db}
}

// GetEventTicketTypeBreakdown returns per-ticket-type metrics for a single event.
// Only confirmed tickets are counted.
func (s *MetricsService) GetEventTicketTypeBreakdown(eventID uuid.UUID) ([]TicketTypeBreakdown, error) {
	type row struct {
		TicketTypeID uuid.UUID
		Name         string
		Quantity     int64
		Sold         int64
		GrossRevenue float64
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			tt.id            AS ticket_type_id,
			tt.name          AS name,
			tt.quantity      AS quantity,
			COUNT(t.id)      AS sold,
			COALESCE(SUM(t.price), 0) AS gross_revenue
		FROM ticket_types tt
		LEFT JOIN tickets t
			ON t.ticket_type_id = tt.id
			AND t.status = 'confirmed'
			AND t.deleted_at IS NULL
		WHERE tt.event_id = ?
		GROUP BY tt.id, tt.name, tt.quantity
		ORDER BY tt.name
	`, eventID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]TicketTypeBreakdown, 0, len(rows))
	for _, r := range rows {
		remaining := r.Quantity - r.Sold
		if remaining < 0 {
			remaining = 0
		}
		capacityPct := 0.0
		if r.Quantity > 0 {
			capacityPct = float64(r.Sold) / float64(r.Quantity) * 100
		}
		result = append(result, TicketTypeBreakdown{
			TicketTypeID: r.TicketTypeID,
			Name:         r.Name,
			Sold:         r.Sold,
			Remaining:    remaining,
			GrossRevenue: r.GrossRevenue,
			CapacityPct:  capacityPct,
		})
	}
	return result, nil
}

// GetEventDailyRevenue returns a zero-filled daily revenue time series for an event.
// The series spans from the earliest confirmed ticket sale date to today.
func (s *MetricsService) GetEventDailyRevenue(eventID uuid.UUID) ([]DailyRevenue, error) {
	type row struct {
		Date    time.Time
		Revenue float64
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			DATE_TRUNC('day', t.created_at) AS date,
			COALESCE(SUM(t.price), 0)       AS revenue
		FROM tickets t
		WHERE t.event_id = ?
		  AND t.status = 'confirmed'
		  AND t.deleted_at IS NULL
		GROUP BY DATE_TRUNC('day', t.created_at)
		ORDER BY date
	`, eventID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []DailyRevenue{}, nil
	}

	// Build a map for quick lookup
	revenueByDate := make(map[string]float64, len(rows))
	for _, r := range rows {
		revenueByDate[r.Date.Format("2006-01-02")] = r.Revenue
	}

	// Zero-fill from first sale date to today
	start := rows[0].Date.Truncate(24 * time.Hour)
	end := time.Now().Truncate(24 * time.Hour)

	var result []DailyRevenue
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		result = append(result, DailyRevenue{
			Date:    key,
			Revenue: revenueByDate[key],
		})
	}
	return result, nil
}

// GetEventCheckInStats returns check-in counts and rate for an event.
func (s *MetricsService) GetEventCheckInStats(eventID uuid.UUID) (CheckInStats, error) {
	type row struct {
		TotalConfirmed int64
		CheckedIn      int64
	}

	var r row
	err := s.db.Raw(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'confirmed')                          AS total_confirmed,
			COUNT(*) FILTER (WHERE status = 'confirmed' AND checked_in_at IS NOT NULL) AS checked_in
		FROM tickets
		WHERE event_id = ?
		  AND deleted_at IS NULL
	`, eventID).Scan(&r).Error
	if err != nil {
		return CheckInStats{}, err
	}

	rate := 0.0
	if r.TotalConfirmed > 0 {
		rate = float64(r.CheckedIn) / float64(r.TotalConfirmed)
	}
	return CheckInStats{
		TotalConfirmed: r.TotalConfirmed,
		CheckedIn:      r.CheckedIn,
		CheckInRate:    rate,
	}, nil
}

// GetPlatformTicketTypeBreakdown returns cross-event ticket-type metrics within a date range.
func (s *MetricsService) GetPlatformTicketTypeBreakdown(start, end time.Time) ([]TicketTypeBreakdown, error) {
	type row struct {
		TicketTypeID uuid.UUID
		Name         string
		Quantity     int64
		Sold         int64
		GrossRevenue float64
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			tt.id            AS ticket_type_id,
			tt.name          AS name,
			tt.quantity      AS quantity,
			COUNT(t.id)      AS sold,
			COALESCE(SUM(t.price), 0) AS gross_revenue
		FROM ticket_types tt
		LEFT JOIN tickets t
			ON t.ticket_type_id = tt.id
			AND t.status = 'confirmed'
			AND t.deleted_at IS NULL
			AND t.created_at BETWEEN ? AND ?
		GROUP BY tt.id, tt.name, tt.quantity
		ORDER BY gross_revenue DESC
	`, start, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]TicketTypeBreakdown, 0, len(rows))
	for _, r := range rows {
		remaining := r.Quantity - r.Sold
		if remaining < 0 {
			remaining = 0
		}
		capacityPct := 0.0
		if r.Quantity > 0 {
			capacityPct = float64(r.Sold) / float64(r.Quantity) * 100
		}
		result = append(result, TicketTypeBreakdown{
			TicketTypeID: r.TicketTypeID,
			Name:         r.Name,
			Sold:         r.Sold,
			Remaining:    remaining,
			GrossRevenue: r.GrossRevenue,
			CapacityPct:  capacityPct,
		})
	}
	return result, nil
}

// GetDailyUserRegistrations returns a daily new-user count time series within a date range.
func (s *MetricsService) GetDailyUserRegistrations(start, end time.Time) ([]DailyCount, error) {
	type row struct {
		Date  time.Time
		Count int64
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			DATE_TRUNC('day', created_at) AS date,
			COUNT(*)                      AS count
		FROM users
		WHERE created_at BETWEEN ? AND ?
		  AND deleted_at IS NULL
		GROUP BY DATE_TRUNC('day', created_at)
		ORDER BY date
	`, start, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []DailyCount{}, nil
	}

	// Build lookup map
	countByDate := make(map[string]int64, len(rows))
	for _, r := range rows {
		countByDate[r.Date.Format("2006-01-02")] = r.Count
	}

	// Zero-fill from start to end
	startDay := start.Truncate(24 * time.Hour)
	endDay := end.Truncate(24 * time.Hour)

	var result []DailyCount
	for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		result = append(result, DailyCount{
			Date:  key,
			Count: countByDate[key],
		})
	}
	return result, nil
}

// GetTopEventsByRevenue returns the top events by gross revenue within a date range.
func (s *MetricsService) GetTopEventsByRevenue(start, end time.Time, limit int) ([]EventRevenue, error) {
	if limit <= 0 {
		limit = 10
	}

	type row struct {
		EventID      uuid.UUID
		Title        string
		GrossRevenue float64
		TicketsSold  int64
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			e.id                      AS event_id,
			e.title                   AS title,
			COALESCE(SUM(t.price), 0) AS gross_revenue,
			COUNT(t.id)               AS tickets_sold
		FROM events e
		LEFT JOIN tickets t
			ON t.event_id = e.id
			AND t.status = 'confirmed'
			AND t.deleted_at IS NULL
			AND t.created_at BETWEEN ? AND ?
		WHERE e.deleted_at IS NULL
		GROUP BY e.id, e.title
		ORDER BY gross_revenue DESC
		LIMIT ?
	`, start, end, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]EventRevenue, 0, len(rows))
	for _, r := range rows {
		result = append(result, EventRevenue{
			EventID:      r.EventID,
			Title:        r.Title,
			GrossRevenue: r.GrossRevenue,
			TicketsSold:  r.TicketsSold,
		})
	}
	return result, nil
}
