package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/models"
	"gorm.io/gorm"
)

// CartItemInput represents a single item in a cart upsert request.
type CartItemInput struct {
	TicketTypeID uuid.UUID `json:"ticket_type_id"`
	Quantity     int       `json:"quantity"`
}

// CartService handles draft cart persistence and validation.
type CartService struct {
	db *gorm.DB
}

// NewCartService creates a new CartService.
func NewCartService(db *gorm.DB) *CartService {
	return &CartService{db: db}
}

// GetCart returns the active draft cart for the user/guest+event, or nil if none exists.
func (s *CartService) GetCart(userID *uuid.UUID, guestSession string, eventID uuid.UUID) (*models.DraftCart, error) {
	var cart models.DraftCart

	query := s.db.
		Preload("Items.TicketType").
		Where("event_id = ? AND status = ? AND expires_at > ?", eventID, models.DraftCartActive, time.Now())

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("guest_session = ?", guestSession)
	}

	err := query.First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

// UpsertCart creates or replaces the active draft cart for the user/guest+event.
// Validates each TicketTypeID exists. Sets ExpiresAt = now+24h. Replaces all items atomically.
// Enforces at most one active cart per user-event pair.
func (s *CartService) UpsertCart(userID *uuid.UUID, guestSession string, eventID uuid.UUID, items []CartItemInput) (*models.DraftCart, error) {
	// Validate all ticket type IDs first (outside transaction for early exit)
	for _, item := range items {
		var tt models.TicketType
		if err := s.db.First(&tt, "id = ?", item.TicketTypeID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("ticket type not found")
			}
			return nil, err
		}
	}

	var cart models.DraftCart
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Find existing active cart
		findQuery := tx.Where("event_id = ? AND status = ?", eventID, models.DraftCartActive)
		if userID != nil {
			findQuery = findQuery.Where("user_id = ?", userID)
		} else {
			findQuery = findQuery.Where("guest_session = ?", guestSession)
		}

		var existing models.DraftCart
		err := findQuery.First(&existing).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new cart
			cart = models.DraftCart{
				UserID:       userID,
				GuestSession: guestSession,
				EventID:      eventID,
				Status:       models.DraftCartActive,
				ExpiresAt:    expiresAt,
			}
			if err := tx.Create(&cart).Error; err != nil {
				return err
			}
		} else {
			// Update existing cart
			existing.ExpiresAt = expiresAt
			existing.UpdatedAt = now
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			// Delete old items
			if err := tx.Where("cart_id = ?", existing.ID).Delete(&models.DraftCartItem{}).Error; err != nil {
				return err
			}
			cart = existing
		}

		// Insert new items
		for _, item := range items {
			cartItem := models.DraftCartItem{
				CartID:       cart.ID,
				TicketTypeID: item.TicketTypeID,
				Quantity:     item.Quantity,
			}
			if err := tx.Create(&cartItem).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload with associations
	if err := s.db.Preload("Items.TicketType").First(&cart, "id = ?", cart.ID).Error; err != nil {
		return nil, err
	}

	return &cart, nil
}

// DeleteCart marks the active cart as completed.
func (s *CartService) DeleteCart(userID *uuid.UUID, guestSession string, eventID uuid.UUID) error {
	query := s.db.Model(&models.DraftCart{}).
		Where("event_id = ? AND status = ?", eventID, models.DraftCartActive)

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("guest_session = ?", guestSession)
	}

	return query.Update("status", models.DraftCartCompleted).Error
}

// ExpireOldCarts marks all active carts with updated_at > 24h ago as expired.
func (s *CartService) ExpireOldCarts() error {
	cutoff := time.Now().Add(-24 * time.Hour)
	return s.db.Model(&models.DraftCart{}).
		Where("status = ? AND updated_at < ?", models.DraftCartActive, cutoff).
		Update("status", models.DraftCartExpired).Error
}

// ValidateAndCleanCart removes items whose ticket type is unavailable.
// Returns the names of removed items.
func (s *CartService) ValidateAndCleanCart(cart *models.DraftCart) ([]string, error) {
	var removed []string
	var validItems []models.DraftCartItem

	for _, item := range cart.Items {
		tt := item.TicketType
		if !tt.IsAvailable() || tt.RemainingTickets() <= 0 {
			removed = append(removed, tt.Name)
			if err := s.db.Delete(&models.DraftCartItem{}, "id = ?", item.ID).Error; err != nil {
				return nil, err
			}
		} else {
			validItems = append(validItems, item)
		}
	}

	cart.Items = validItems
	return removed, nil
}
