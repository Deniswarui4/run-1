package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/models"
	"github.com/warui/event-ticketing-api/internal/services"
	"gorm.io/gorm"
)

// GuestHandler handles guest checkout and ticket lookup.
type GuestHandler struct {
	db             *gorm.DB
	cfg            *config.Config
	pesapalService *services.PesapalService
	storageService *services.StorageService
	qrcodeService  *services.QRCodeService
	pdfService     *services.PDFService
	emailService   *services.EmailService
}

// NewGuestHandler creates a new GuestHandler.
func NewGuestHandler(
	db *gorm.DB,
	cfg *config.Config,
	pesapalService *services.PesapalService,
	storageService *services.StorageService,
	qrcodeService *services.QRCodeService,
	pdfService *services.PDFService,
	emailService *services.EmailService,
) *GuestHandler {
	return &GuestHandler{
		db:             db,
		cfg:            cfg,
		pesapalService: pesapalService,
		storageService: storageService,
		qrcodeService:  qrcodeService,
		pdfService:     pdfService,
		emailService:   emailService,
	}
}

// generateGuestToken creates a cryptographically random opaque token.
func generateGuestToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// InitiateGuestCheckout handles POST /api/v1/guest/checkout
func (h *GuestHandler) InitiateGuestCheckout(c *gin.Context) {
	var req struct {
		Email  string `json:"email" binding:"required"`
		EventID string `json:"event_id" binding:"required"`
		Items  []struct {
			TicketTypeID string `json:"ticket_type_id" binding:"required"`
			Quantity     int    `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required,min=1,dive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email
	if !services.ValidateGuestEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
		return
	}

	// Get event
	var event models.Event
	if err := h.db.First(&event, "id = ?", req.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	// Validate ticket types and calculate total
	var totalAmount float64
	var ticketItems []map[string]interface{}

	for _, item := range req.Items {
		var ticketType models.TicketType
		if err := h.db.First(&ticketType, "id = ? AND event_id = ?", item.TicketTypeID, req.EventID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket type not found"})
			return
		}

		if !ticketType.IsAvailable() {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("tickets for %s not available for sale", ticketType.Name)})
			return
		}

		if item.Quantity > ticketType.MaxPerOrder {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("maximum %d tickets per order for %s", ticketType.MaxPerOrder, ticketType.Name)})
			return
		}

		totalAmount += ticketType.Price * float64(item.Quantity)
		ticketItems = append(ticketItems, map[string]interface{}{
			"ticket_type_id": item.TicketTypeID,
			"quantity":       item.Quantity,
			"price":          ticketType.Price,
			"name":           ticketType.Name,
		})
	}

	// Platform fee
	var settings models.PlatformSettings
	h.db.First(&settings)
	platformFee := totalAmount * (settings.PlatformFeePercentage / 100)

	// Prepare metadata to store in transaction
	metadataMap := map[string]interface{}{
		"guest_email": req.Email,
		"items":       ticketItems,
		"email":       req.Email,
	}
	metadataJSON, _ := json.Marshal(metadataMap)
	metadataStr := string(metadataJSON)

	// Create pending transaction with nil user_id for guest purchases
	transaction := &models.Transaction{
		UserID:           nil, // Guest purchase - no user ID
		EventID:          &event.ID,
		Type:             models.TransactionTypeTicketPurchase,
		Status:           models.TransactionStatusPending,
		Amount:           totalAmount,
		Currency:         h.cfg.Currency,
		PlatformFee:      platformFee,
		NetAmount:        totalAmount - platformFee,
		PaymentGateway:   "pesapal",
		PaymentReference: fmt.Sprintf("TXN-%s-%d", uuid.New().String()[:8], time.Now().Unix()),
		Description:      fmt.Sprintf("Guest purchase of tickets for %s", event.Title),
		PaymentMetadata:  &metadataStr,
	}

	if err := h.db.Create(transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
		return
	}

	// Initialize Pesapal payment
	metadata := map[string]interface{}{
		"transaction_id": transaction.ID.String(),
		"event_id":       event.ID.String(),
		"guest_email":    req.Email,
		"items":          ticketItems,
		"first_name":     "Guest",
		"last_name":      "",
	}

	paymentInit, err := h.pesapalService.InitializeTransaction(
		req.Email,
		totalAmount,
		transaction.PaymentReference,
		metadata,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "payment initialization failed"})
		return
	}

	// Issue guest_session HTTP-only cookie (30-day expiry)
	token, err := generateGuestToken()
	if err == nil {
		guestSession := &models.GuestSession{
			Token:     token,
			Email:     req.Email,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}
		if dbErr := h.db.Create(guestSession).Error; dbErr == nil {
			c.SetCookie(
				"guest_session",
				token,
				int((30 * 24 * time.Hour).Seconds()),
				"/",
				"",
				true,  // secure
				true,  // httpOnly
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id":    transaction.ID,
		"payment_reference": transaction.PaymentReference,
		"authorization_url": paymentInit.RedirectURL,
		"amount":            totalAmount,
		"currency":          h.cfg.Currency,
	})
}

// LookupGuestTickets handles GET /api/v1/guest/tickets?email=&reference=
func (h *GuestHandler) LookupGuestTickets(c *gin.Context) {
	email := c.Query("email")
	reference := c.Query("reference")

	if email == "" || reference == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and reference are required"})
		return
	}

	if !services.ValidateGuestEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
		return
	}

	// Find transaction by reference
	var transaction models.Transaction
	if err := h.db.First(&transaction, "payment_reference = ?", reference).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	// Verify guest_email matches by checking tickets
	var tickets []models.Ticket
	if err := h.db.Preload("Event").Preload("TicketType").
		Where("transaction_id = ? AND guest_email = ?", transaction.ID, email).
		Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tickets"})
		return
	}

	if len(tickets) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no tickets found for this email and reference"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}
