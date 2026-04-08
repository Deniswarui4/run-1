// Feature: guest-checkout-cart-sharing-metrics
package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrEventNotFound is returned when the event does not exist.
var ErrEventNotFound = errors.New("event not found")

// ErrEventNotPublished is returned when a non-published event is targeted for featuring.
var ErrEventNotPublished = errors.New("only published events can be featured")

// FeatureService manages featured-event logic.
type FeatureService struct {
	db *gorm.DB
}

// NewFeatureService creates a new FeatureService.
func NewFeatureService(db *gorm.DB) *FeatureService {
	return &FeatureService{db: db}
}

// ToggleFeatured toggles the is_featured flag for a published event.
// When manual=true the featured_type is set to "manual" (on) or "none" (off).
// Returns ErrEventNotFound or ErrEventNotPublished on failure.
func (s *FeatureService) ToggleFeatured(eventID uuid.UUID, manual bool) (*models.Event, error) {
	var event models.Event
	if err := s.db.First(&event, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	if event.Status != models.EventStatusPublished {
		return nil, ErrEventNotPublished
	}

	if manual {
		// Toggle: flip is_featured
		event.IsFeatured = !event.IsFeatured
		if event.IsFeatured {
			event.FeaturedType = models.FeaturedTypeManual
		} else {
			event.FeaturedType = models.FeaturedTypeNone
		}
	} else {
		// Programmatic unfeature
		event.IsFeatured = false
		event.FeaturedType = models.FeaturedTypeNone
	}

	if err := s.db.Save(&event).Error; err != nil {
		return nil, err
	}

	// Reload with associations for the response
	if err := s.db.Preload("Organizer").Preload("TicketTypes").First(&event, event.ID).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

// GetSalesVelocity counts confirmed tickets created in the last 24 hours for the event.
func (s *FeatureService) GetSalesVelocity(eventID uuid.UUID) (int64, error) {
	var count int64
	since := time.Now().Add(-24 * time.Hour)
	err := s.db.Model(&models.Ticket{}).
		Where("event_id = ? AND status = ? AND created_at >= ? AND deleted_at IS NULL",
			eventID, models.TicketStatusConfirmed, since).
		Count(&count).Error
	return count, err
}

// EvaluateAutoFeature iterates all published events and applies auto-feature logic.
// It is intended to be called hourly by a background goroutine.
func (s *FeatureService) EvaluateAutoFeature() error {
	// Fetch the current threshold from platform settings
	var settings models.PlatformSettings
	if err := s.db.First(&settings).Error; err != nil {
		return err
	}

	// If threshold is 0, skip evaluation entirely
	if settings.AutoFeatureThreshold == 0 {
		return nil
	}
	threshold := int64(settings.AutoFeatureThreshold)

	// Fetch all published events
	var events []models.Event
	if err := s.db.Where("status = ?", models.EventStatusPublished).Find(&events).Error; err != nil {
		return err
	}

	now := time.Now()

	for _, event := range events {
		velocity, err := s.GetSalesVelocity(event.ID)
		if err != nil {
			// Log and continue rather than aborting the whole evaluation
			continue
		}

		if velocity >= threshold {
			// Only auto-feature if not already manually featured
			if event.FeaturedType != models.FeaturedTypeManual {
				updates := map[string]interface{}{
					"is_featured":   true,
					"featured_type": models.FeaturedTypeAuto,
				}
				s.db.Model(&event).Updates(updates)

				// Reset consecutive_below_threshold
				eval := models.EventFeatureEvaluation{
					EventID:                   event.ID,
					ConsecutiveBelowThreshold: 0,
					LastEvaluatedAt:           &now,
				}
				s.db.Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "event_id"}},
					DoUpdates: clause.Assignments(map[string]interface{}{
						"consecutive_below_threshold": 0,
						"last_evaluated_at":           now,
					}),
				}).Create(&eval)
			}
		} else {
			// Only apply unfeature logic to auto-featured events; skip manually-featured
			if event.FeaturedType == models.FeaturedTypeAuto {
				// Upsert evaluation record, incrementing the counter
				var eval models.EventFeatureEvaluation
				err := s.db.Where("event_id = ?", event.ID).First(&eval).Error
				if err != nil {
					// No record yet — create with count = 1
					eval = models.EventFeatureEvaluation{
						EventID:                   event.ID,
						ConsecutiveBelowThreshold: 1,
						LastEvaluatedAt:           &now,
					}
					s.db.Create(&eval)
				} else {
					eval.ConsecutiveBelowThreshold++
					eval.LastEvaluatedAt = &now
					s.db.Save(&eval)
				}

				if eval.ConsecutiveBelowThreshold >= 3 {
					updates := map[string]interface{}{
						"is_featured":   false,
						"featured_type": models.FeaturedTypeNone,
					}
					s.db.Model(&event).Updates(updates)
				}
			}
			// Manually-featured events are skipped for unfeature logic
		}
	}

	return nil
}
