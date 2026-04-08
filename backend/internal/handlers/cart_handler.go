package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/middleware"
	"github.com/warui/event-ticketing-api/internal/services"
	"gorm.io/gorm"
)

// CartHandler handles cart-related HTTP requests.
type CartHandler struct {
	db          *gorm.DB
	cfg         *config.Config
	cartService *services.CartService
}

// NewCartHandler creates a new CartHandler.
func NewCartHandler(db *gorm.DB, cfg *config.Config, cartService *services.CartService) *CartHandler {
	return &CartHandler{db: db, cfg: cfg, cartService: cartService}
}

// resolveIdentity extracts userID from JWT context or guest_session cookie.
func (h *CartHandler) resolveIdentity(c *gin.Context) (userID *uuid.UUID, guestSession string) {
	if id, err := middleware.GetUserID(c); err == nil {
		userID = &id
		return
	}
	guestSession = c.GetHeader("X-Guest-Session")
	if guestSession == "" {
		guestSession, _ = c.Cookie("guest_session")
	}
	return
}

// GetCart handles GET /api/v1/cart?event_id=
func (h *CartHandler) GetCart(c *gin.Context) {
	eventIDStr := c.Query("event_id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	userID, guestSession := h.resolveIdentity(c)

	cart, err := h.cartService.GetCart(userID, guestSession, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch cart"})
		return
	}

	if cart == nil {
		c.JSON(http.StatusOK, gin.H{"items": []interface{}{}})
		return
	}

	removed, err := h.cartService.ValidateAndCleanCart(cart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate cart"})
		return
	}

	if len(removed) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"id":            cart.ID,
			"event_id":      cart.EventID,
			"status":        cart.Status,
			"expires_at":    cart.ExpiresAt,
			"items":         cart.Items,
			"removed_items": removed,
		})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// UpsertCart handles PUT /api/v1/cart
func (h *CartHandler) UpsertCart(c *gin.Context) {
	var req struct {
		EventID uuid.UUID `json:"event_id" binding:"required"`
		Items   []struct {
			TicketTypeID uuid.UUID `json:"ticket_type_id" binding:"required"`
			Quantity     int       `json:"quantity" binding:"required"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, item := range req.Items {
		if item.Quantity < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be at least 1"})
			return
		}
	}

	userID, guestSession := h.resolveIdentity(c)
	if userID == nil && guestSession == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authentication or guest session required"})
		return
	}

	cartItems := make([]services.CartItemInput, len(req.Items))
	for i, item := range req.Items {
		cartItems[i] = services.CartItemInput{
			TicketTypeID: item.TicketTypeID,
			Quantity:     item.Quantity,
		}
	}

	cart, err := h.cartService.UpsertCart(userID, guestSession, req.EventID, cartItems)
	if err != nil {
		if err.Error() == "ticket type not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ticket type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save cart"})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// DeleteCart handles DELETE /api/v1/cart?event_id=
func (h *CartHandler) DeleteCart(c *gin.Context) {
	eventIDStr := c.Query("event_id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	userID, guestSession := h.resolveIdentity(c)

	if err := h.cartService.DeleteCart(userID, guestSession, eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cart cleared"})
}
