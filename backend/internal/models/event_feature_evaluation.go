package models

import (
	"time"

	"github.com/google/uuid"
)

// EventFeatureEvaluation tracks the auto-feature evaluation state for an event.
type EventFeatureEvaluation struct {
	EventID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"event_id"`
	ConsecutiveBelowThreshold int        `gorm:"not null;default:0" json:"consecutive_below_threshold"`
	LastEvaluatedAt           *time.Time `json:"last_evaluated_at,omitempty"`
}
