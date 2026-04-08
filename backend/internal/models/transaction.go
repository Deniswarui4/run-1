package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionStatus string
type TransactionType string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusRefunded  TransactionStatus = "refunded"

	TransactionTypeTicketPurchase TransactionType = "ticket_purchase"
	TransactionTypeRefund         TransactionType = "refund"
	TransactionTypeWithdrawal     TransactionType = "withdrawal"
)

type Transaction struct {
	ID          uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      *uuid.UUID        `gorm:"type:uuid" json:"user_id,omitempty"` // Nullable for guest purchases
	EventID     *uuid.UUID        `gorm:"type:uuid" json:"event_id,omitempty"`
	Type        TransactionType   `gorm:"type:varchar(30);not null" json:"type"`
	Status      TransactionStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Amount      float64           `gorm:"not null" json:"amount"`
	Currency    string            `gorm:"default:'NGN'" json:"currency"`
	PlatformFee float64           `gorm:"default:0" json:"platform_fee"`
	NetAmount   float64           `gorm:"not null" json:"net_amount"`

	// Payment gateway details
	PaymentGateway   string  `json:"payment_gateway"`
	PaymentReference string  `gorm:"uniqueIndex" json:"payment_reference"`
	PaymentMetadata  *string `gorm:"type:jsonb" json:"payment_metadata,omitempty"`

	Description   string `json:"description"`
	FailureReason string `json:"failure_reason,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Event   *Event   `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Tickets []Ticket `gorm:"foreignKey:TransactionID" json:"tickets,omitempty"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
