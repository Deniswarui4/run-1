package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleOrganizer Role = "organizer"
	RoleAttendee  Role = "attendee"
)

type User struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email      string         `gorm:"uniqueIndex;not null" json:"email"`
	Password   string         `gorm:"not null" json:"-"`
	FirstName  string         `gorm:"not null" json:"first_name"`
	LastName   string         `gorm:"not null" json:"last_name"`
	Phone      string         `json:"phone"`
	Role                 Role           `gorm:"type:varchar(20);not null;default:'attendee'" json:"role"`
	IsActive             bool           `gorm:"default:true" json:"is_active"`
	IsVerified           bool           `gorm:"default:false" json:"is_verified"`
	VerificationToken    *string        `gorm:"index" json:"-"`
	VerificationExpiry   *time.Time     `json:"-"`
	PasswordResetToken   *string        `gorm:"index" json:"-"`
	PasswordResetExpiry  *time.Time     `json:"-"`
	TwoFactorSecret      *string        `json:"-"`
	TwoFactorEnabled     bool           `gorm:"default:false" json:"two_factor_enabled"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Events       []Event       `gorm:"foreignKey:OrganizerID" json:"events,omitempty"`
	Tickets      []Ticket      `gorm:"foreignKey:AttendeeID" json:"tickets,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:UserID" json:"transactions,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// HasPermission checks if user has required role
func (u *User) HasPermission(requiredRole Role) bool {
	roleHierarchy := map[Role]int{
		RoleAttendee:  1,
		RoleOrganizer: 2,
		RoleModerator: 3,
		RoleAdmin:     4,
	}
	return roleHierarchy[u.Role] >= roleHierarchy[requiredRole]
}

// IsAdmin checks if user is admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsModerator checks if user is moderator or admin
func (u *User) IsModerator() bool {
	return u.Role == RoleModerator || u.Role == RoleAdmin
}

// IsOrganizer checks if user is organizer, moderator, or admin
func (u *User) IsOrganizer() bool {
	return u.Role == RoleOrganizer || u.Role == RoleModerator || u.Role == RoleAdmin
}
