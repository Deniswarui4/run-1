package models

import (
	"time"

	"github.com/google/uuid"
)

// DraftCartStatus represents the lifecycle state of a draft cart.
type DraftCartStatus string

const (
	DraftCartActive    DraftCartStatus = "active"
	DraftCartExpired   DraftCartStatus = "expired"
	DraftCartCompleted DraftCartStatus = "completed"
)

// DraftCart holds a user's or guest's in-progress ticket selections for an event.
type DraftCart struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       *uuid.UUID      `gorm:"type:uuid" json:"user_id,omitempty"`
	GuestSession string          `json:"guest_session,omitempty"`
	EventID      uuid.UUID       `gorm:"type:uuid;not null" json:"event_id"`
	Status       DraftCartStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	ExpiresAt    time.Time       `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`

	// Relationships
	Items []DraftCartItem `gorm:"foreignKey:CartID" json:"items,omitempty"`
}

// DraftCartItem is a single ticket-type line in a draft cart.
type DraftCartItem struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CartID       uuid.UUID  `gorm:"type:uuid;not null" json:"cart_id"`
	TicketTypeID uuid.UUID  `gorm:"type:uuid;not null" json:"ticket_type_id"`
	Quantity     int        `gorm:"not null" json:"quantity"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relationships
	TicketType TicketType `gorm:"foreignKey:TicketTypeID" json:"ticket_type,omitempty"`
}
