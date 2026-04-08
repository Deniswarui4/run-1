package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlatformSettings stores platform-wide configuration
type PlatformSettings struct {
	ID                      uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PlatformFeePercentage   float64   `gorm:"not null;default:5.0" json:"platform_fee_percentage"`
	WithdrawalFeePercentage float64   `gorm:"not null;default:2.5" json:"withdrawal_fee_percentage"`
	MinWithdrawalAmount     float64   `gorm:"default:1000" json:"min_withdrawal_amount"`
	Currency                string    `gorm:"default:'NGN'" json:"currency"`
	AutoFeatureThreshold    int       `gorm:"not null;default:0" json:"auto_feature_threshold"`
	UpdatedBy               uuid.UUID `gorm:"type:uuid" json:"updated_by"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func (p *PlatformSettings) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type WithdrawalStatus string

const (
	WithdrawalStatusPending   WithdrawalStatus = "pending"
	WithdrawalStatusApproved  WithdrawalStatus = "approved"
	WithdrawalStatusRejected  WithdrawalStatus = "rejected"
	WithdrawalStatusProcessed WithdrawalStatus = "processed"
)

// WithdrawalRequest represents organizer withdrawal requests
type WithdrawalRequest struct {
	ID            uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizerID   uuid.UUID        `gorm:"type:uuid;not null" json:"organizer_id"`
	Amount        float64          `gorm:"not null" json:"amount"`
	WithdrawalFee float64          `gorm:"not null" json:"withdrawal_fee"`
	NetAmount     float64          `gorm:"not null" json:"net_amount"`
	Status        WithdrawalStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`

	// Bank details
	BankName      string `gorm:"not null" json:"bank_name"`
	AccountNumber string `gorm:"not null" json:"account_number"`
	AccountName   string `gorm:"not null" json:"account_name"`

	// Admin review
	ReviewedBy    *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewComment string     `gorm:"type:text" json:"review_comment,omitempty"`

	// Processing
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	TransactionRef string     `json:"transaction_ref,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Organizer User  `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`
	Reviewer  *User `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

func (w *WithdrawalRequest) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// OrganizerBalance tracks organizer earnings
type OrganizerBalance struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizerID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"organizer_id"`
	TotalEarnings    float64   `gorm:"default:0" json:"total_earnings"`
	AvailableBalance float64   `gorm:"default:0" json:"available_balance"`
	PendingBalance   float64   `gorm:"default:0" json:"pending_balance"`
	WithdrawnAmount  float64   `gorm:"default:0" json:"withdrawn_amount"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relationships
	Organizer User `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`
}

func (o *OrganizerBalance) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
