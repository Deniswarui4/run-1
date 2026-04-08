package models

import (
	"time"

	"github.com/google/uuid"
)

// GuestSession tracks guest identity via an opaque token stored in an HTTP-only cookie.
type GuestSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	Email     string    `gorm:"not null" json:"email"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
