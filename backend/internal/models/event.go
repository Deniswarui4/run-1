package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPending   EventStatus = "pending"
	EventStatusApproved  EventStatus = "approved"
	EventStatusRejected  EventStatus = "rejected"
	EventStatusPublished EventStatus = "published"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusCompleted EventStatus = "completed"
)

// FeaturedType distinguishes how an event was featured.
type FeaturedType = string

const (
	FeaturedTypeNone   FeaturedType = "none"
	FeaturedTypeManual FeaturedType = "manual"
	FeaturedTypeAuto   FeaturedType = "auto"
)

type Event struct {
	ID          uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string      `gorm:"not null" json:"title"`
	Description string      `gorm:"type:text" json:"description"`
	Category    string      `json:"category"`
	Venue       string      `gorm:"not null" json:"venue"`
	Address     string      `json:"address"`
	City        string      `json:"city"`
	Country     string      `json:"country"`
	ImageURL    string      `json:"image_url"`
	VideoURL    string      `json:"video_url"` // Promotional video URL
	StartDate   time.Time   `gorm:"not null" json:"start_date"`
	EndDate     time.Time   `gorm:"not null" json:"end_date"`
	Status      EventStatus `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	IsFeatured  bool        `gorm:"default:false" json:"is_featured"`
	FeaturedType string     `gorm:"type:varchar(10);not null;default:'none'" json:"featured_type"`

	OrganizerID uuid.UUID `gorm:"type:uuid;not null" json:"organizer_id"`
	Organizer   User      `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`

	// Moderation
	ModeratorID       *uuid.UUID `gorm:"type:uuid" json:"moderator_id,omitempty"`
	Moderator         *User      `gorm:"foreignKey:ModeratorID" json:"moderator,omitempty"`
	ModerationComment string     `gorm:"type:text" json:"moderation_comment,omitempty"`
	ModeratedAt       *time.Time `json:"moderated_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	TicketTypes []TicketType `gorm:"foreignKey:EventID" json:"ticket_types,omitempty"`
	Tickets     []Ticket     `gorm:"foreignKey:EventID" json:"tickets,omitempty"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type TicketType struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventID     uuid.UUID `gorm:"type:uuid;not null" json:"event_id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"not null" json:"price"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	Sold        int       `gorm:"default:0" json:"sold"`
	MaxPerOrder int       `gorm:"default:10" json:"max_per_order"`
	SaleStart   time.Time `json:"sale_start"`
	SaleEnd     time.Time `json:"sale_end"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Event   Event    `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Tickets []Ticket `gorm:"foreignKey:TicketTypeID" json:"tickets,omitempty"`
}

func (t *TicketType) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t *TicketType) IsAvailable() bool {
	now := time.Now()
	return t.IsActive &&
		t.Sold < t.Quantity &&
		!now.Before(t.SaleStart) && // now >= SaleStart
		!now.After(t.SaleEnd) // now <= SaleEnd
}

func (t *TicketType) RemainingTickets() int {
	return t.Quantity - t.Sold
}
