package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStatus string

const (
	TicketStatusPending   TicketStatus = "pending"
	TicketStatusConfirmed TicketStatus = "confirmed"
	TicketStatusCancelled TicketStatus = "cancelled"
	TicketStatusUsed      TicketStatus = "used"
)

type Ticket struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TicketNumber  string     `gorm:"uniqueIndex;not null" json:"ticket_number"`
	EventID       uuid.UUID  `gorm:"type:uuid;not null" json:"event_id"`
	TicketTypeID  uuid.UUID  `gorm:"type:uuid;not null" json:"ticket_type_id"`
	AttendeeID    *uuid.UUID `gorm:"type:uuid" json:"attendee_id,omitempty"`
	TransactionID uuid.UUID  `gorm:"type:uuid;not null" json:"transaction_id"`

	// GuestEmail is set for guest purchases; AttendeeID is nil in that case.
	GuestEmail *string `json:"guest_email,omitempty"`

	Status    TicketStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Price     float64      `gorm:"not null" json:"price"`
	QRCodeURL string       `json:"qr_code_url"`
	PDFURL    string       `json:"pdf_url"`

	CheckedInAt *time.Time `json:"checked_in_at,omitempty"`
	CheckedInBy *uuid.UUID `gorm:"type:uuid" json:"checked_in_by,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Event       Event       `gorm:"foreignKey:EventID" json:"event,omitempty"`
	TicketType  TicketType  `gorm:"foreignKey:TicketTypeID" json:"ticket_type,omitempty"`
	Attendee    *User       `gorm:"foreignKey:AttendeeID" json:"attendee,omitempty"`
	Transaction Transaction `gorm:"foreignKey:TransactionID" json:"transaction,omitempty"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.TicketNumber == "" {
		t.TicketNumber = generateTicketNumber()
	}
	return nil
}

func generateTicketNumber() string {
	return "TKT-" + uuid.New().String()[:8]
}
