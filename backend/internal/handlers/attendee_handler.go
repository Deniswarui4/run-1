package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/middleware"
	"github.com/warui/event-ticketing-api/internal/models"
	"github.com/warui/event-ticketing-api/internal/services"
	"gorm.io/gorm"
)

type AttendeeHandler struct {
	db             *gorm.DB
	cfg            *config.Config
	pesapalService *services.PesapalService
	storageService *services.StorageService
	qrcodeService  *services.QRCodeService
	pdfService     *services.PDFService
	emailService   *services.EmailService
}

func NewAttendeeHandler(
	db *gorm.DB,
	cfg *config.Config,
	pesapalService *services.PesapalService,
	storageService *services.StorageService,
	qrcodeService *services.QRCodeService,
	pdfService *services.PDFService,
	emailService *services.EmailService,
) *AttendeeHandler {
	return &AttendeeHandler{
		db:             db,
		cfg:            cfg,
		pesapalService: pesapalService,
		storageService: storageService,
		qrcodeService:  qrcodeService,
		pdfService:     pdfService,
		emailService:   emailService,
	}
}

// GetPublishedEvents retrieves all published events
func (h *AttendeeHandler) GetPublishedEvents(c *gin.Context) {
	category := c.Query("category")
	city := c.Query("city")
	search := c.Query("search")

	query := h.db.Preload("Organizer").Preload("TicketTypes").Where("status = ?", models.EventStatusPublished)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if city != "" {
		query = query.Where("city = ?", city)
	}

	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Show ongoing and future events (not ended yet)
	query = query.Where("end_date > ?", time.Now())

	var events []models.Event
	if err := query.Order("start_date ASC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetEventDetails retrieves details of a specific event
func (h *AttendeeHandler) GetEventDetails(c *gin.Context) {
	eventID := c.Param("id")

	var event models.Event
	if err := h.db.Preload("Organizer").Preload("TicketTypes").First(&event, "id = ? AND status = ?", eventID, models.EventStatusPublished).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// InitiateTicketPurchase initiates a ticket purchase
func (h *AttendeeHandler) InitiateTicketPurchase(c *gin.Context) {
	attendeeID, _ := middleware.GetUserID(c)

	var req struct {
		EventID string `json:"event_id" binding:"required"`
		Items   []struct {
			TicketTypeID string `json:"ticket_type_id" binding:"required"`
			Quantity     int    `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required,min=1,dive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get event
	var event models.Event
	if err := h.db.First(&event, "id = ?", req.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Validate all ticket types and calculate total
	var totalAmount float64
	var ticketItems []map[string]interface{}

	for _, item := range req.Items {
		// Get ticket type
		var ticketType models.TicketType
		if err := h.db.First(&ticketType, "id = ? AND event_id = ?", item.TicketTypeID, req.EventID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Ticket type %s not found", item.TicketTypeID)})
			return
		}

		// Check availability
		if !ticketType.IsAvailable() {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Tickets for %s not available for sale", ticketType.Name)})
			return
		}

		// Check quantity
		if item.Quantity > ticketType.RemainingTickets() {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Not enough tickets available for %s", ticketType.Name)})
			return
		}

		if item.Quantity > ticketType.MaxPerOrder {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Maximum %d tickets per order for %s", ticketType.MaxPerOrder, ticketType.Name)})
			return
		}

		// Add to total
		itemTotal := ticketType.Price * float64(item.Quantity)
		totalAmount += itemTotal

		// Store item info for metadata
		ticketItems = append(ticketItems, map[string]interface{}{
			"ticket_type_id": item.TicketTypeID,
			"quantity":       item.Quantity,
			"price":          ticketType.Price,
			"name":           ticketType.Name,
		})
	}

	// Get platform settings for fee calculation
	var settings models.PlatformSettings
	h.db.First(&settings)
	platformFee := totalAmount * (settings.PlatformFeePercentage / 100)

	// Get user
	var user models.User
	h.db.First(&user, attendeeID)

	// Prepare metadata to store in transaction
	metadataMap := map[string]interface{}{
		"attendee_id": attendeeID.String(),
		"items":       ticketItems,
		"email":       user.Email,
		"first_name":  user.FirstName,
		"last_name":   user.LastName,
	}
	metadataJSON, _ := json.Marshal(metadataMap)
	metadataStr := string(metadataJSON)

	// Create transaction
	transaction := &models.Transaction{
		UserID:           &attendeeID, // Convert to pointer
		EventID:          &event.ID,
		Type:             models.TransactionTypeTicketPurchase,
		Status:           models.TransactionStatusPending,
		Amount:           totalAmount,
		Currency:         h.cfg.Currency,
		PlatformFee:      platformFee,
		NetAmount:        totalAmount - platformFee,
		PaymentGateway:   "pesapal",
		PaymentReference: fmt.Sprintf("TXN-%s-%d", uuid.New().String()[:8], time.Now().Unix()),
		Description:      fmt.Sprintf("Purchase of tickets for %s", event.Title),
		PaymentMetadata:  &metadataStr,
	}

	if err := h.db.Create(transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Initialize payment with Pesapal
	metadata := map[string]interface{}{
		"transaction_id": transaction.ID.String(),
		"event_id":       event.ID.String(),
		"attendee_id":    attendeeID.String(),
		"items":          ticketItems, // Store all cart items
		"first_name":     user.FirstName,
		"last_name":      user.LastName,
	}

	paymentInit, err := h.pesapalService.InitializeTransaction(
		user.Email,
		totalAmount,
		transaction.PaymentReference,
		metadata,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id":    transaction.ID,
		"payment_reference": transaction.PaymentReference,
		"authorization_url": paymentInit.RedirectURL,
		"amount":            totalAmount,
		"currency":          h.cfg.Currency,
	})
}

// VerifyPayment verifies a payment and creates tickets
func (h *AttendeeHandler) VerifyPayment(c *gin.Context) {
	// Pesapal sends OrderTrackingId as query parameter
	orderTrackingID := c.Query("OrderTrackingId")
	if orderTrackingID == "" {
		// Fallback to reference for backward compatibility
		orderTrackingID = c.Query("reference")
	}
	
	if orderTrackingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order tracking ID required"})
		return
	}

	// Verify with Pesapal
	verification, err := h.pesapalService.VerifyTransaction(orderTrackingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment verification failed: " + err.Error()})
		return
	}

	// Get transaction by merchant reference
	var transaction models.Transaction
	if err := h.db.First(&transaction, "payment_reference = ?", verification.MerchantReference).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Check if already processed
	if transaction.Status == models.TransactionStatusCompleted {
		// Get existing tickets for this transaction
		var existingTickets []models.Ticket
		h.db.Preload("Event").Preload("TicketType").Where("transaction_id = ?", transaction.ID).Find(&existingTickets)

		c.JSON(http.StatusOK, gin.H{
			"message": "Payment already verified",
			"status":  "success",
			"tickets": existingTickets,
		})
		return
	}

	// Verify with Pesapal
	if !h.pesapalService.IsTransactionSuccessful(verification) {
		transaction.Status = models.TransactionStatusFailed
		transaction.FailureReason = verification.PaymentStatusDescription
		h.db.Save(&transaction)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment was not successful"})
		return
	}

	// Extract metadata - Pesapal doesn't return metadata, so we get it from transaction
	eventID := transaction.EventID.String()

	// Get event
	var event models.Event
	h.db.First(&event, "id = ?", eventID)

	// Parse metadata from transaction
	var metadata map[string]interface{}
	var cartItems []map[string]interface{}
	var guestEmail string
	isGuest := transaction.UserID == nil

	if transaction.PaymentMetadata != nil {
		if err := json.Unmarshal([]byte(*transaction.PaymentMetadata), &metadata); err == nil {
			// Extract cart items
			if items, ok := metadata["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						cartItems = append(cartItems, itemMap)
					}
				}
			}
			// Extract guest email if present
			if email, ok := metadata["email"].(string); ok {
				guestEmail = email
			}
		}
	}

	// Create tickets for all items in cart
	var tickets []models.Ticket
	var user models.User
	if !isGuest && transaction.UserID != nil {
		h.db.First(&user, *transaction.UserID)
	}

	for _, item := range cartItems {
		ticketTypeID, _ := item["ticket_type_id"].(string)

		// Handle quantity - could be float64 or string
		var quantity int
		switch v := item["quantity"].(type) {
		case float64:
			quantity = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				quantity = parsed
			} else {
				quantity = 1
			}
		default:
			quantity = 1
		}

		// Get ticket type
		var ticketType models.TicketType
		h.db.First(&ticketType, "id = ?", ticketTypeID)

		// Create tickets for this item
		for i := 0; i < quantity; i++ {
			ticket := models.Ticket{
				EventID:       event.ID,
				TicketTypeID:  ticketType.ID,
				TransactionID: transaction.ID,
				Status:        models.TicketStatusConfirmed,
				Price:         ticketType.Price,
			}

			if isGuest {
				// Guest ticket: set guest_email, leave attendee_id nil
				ticket.GuestEmail = &guestEmail
			} else {
				// Authenticated ticket: set attendee_id
				ticket.AttendeeID = transaction.UserID
			}

			if err := h.db.Create(&ticket).Error; err != nil {
				continue
			}

			// Generate QR code
			var qrData []byte
			var err error
			qrData, err = h.qrcodeService.GenerateTicketQRCode(ticket.TicketNumber, ticket.ID.String())
			if err == nil {
				qrFilename := services.GenerateUniqueFilename(fmt.Sprintf("qr-%s", ticket.TicketNumber), "png")
				qrURL, _ := h.storageService.UploadFile(qrData, "tickets/qrcodes", qrFilename)
				ticket.QRCodeURL = qrURL
			}

			// Generate PDF
			ticket.Event = event
			ticket.TicketType = ticketType

			var pdfData []byte
			if isGuest {
				// Build a minimal user-like struct for the PDF generator
				guestUser := &models.User{
					Email:     guestEmail,
					FirstName: "Guest",
					LastName:  "",
				}
				pdfData, err = h.pdfService.GenerateTicketPDF(&ticket, &event, guestUser, qrData)
			} else {
				pdfData, err = h.pdfService.GenerateTicketPDF(&ticket, &event, &user, qrData)
			}
			if err == nil {
				pdfFilename := services.GenerateUniqueFilename(fmt.Sprintf("ticket-%s", ticket.TicketNumber), "pdf")
				pdfURL, _ := h.storageService.UploadFile(pdfData, "tickets/pdfs", pdfFilename)
				ticket.PDFURL = pdfURL
			}

			h.db.Save(&ticket)
			tickets = append(tickets, ticket)

			// Send email with PDF attachment
			if isGuest {
				guestUser := &models.User{
					Email:     guestEmail,
					FirstName: "Guest",
					LastName:  "",
				}
				go h.emailService.SendTicketEmail(&ticket, &event, guestUser, pdfData)
			} else {
				go h.emailService.SendTicketEmail(&ticket, &event, &user, pdfData)
			}
		}

		// Update ticket type sold count
		ticketType.Sold += quantity
		h.db.Save(&ticketType)
	}

	// Update transaction status
	transaction.Status = models.TransactionStatusCompleted
	h.db.Save(&transaction)

	// Update organizer balance
	var balance models.OrganizerBalance
	if err := h.db.Where("organizer_id = ?", event.OrganizerID).First(&balance).Error; err == nil {
		balance.TotalEarnings += transaction.NetAmount
		balance.AvailableBalance += transaction.NetAmount
		h.db.Save(&balance)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment verified successfully",
		"status":  "success",
		"tickets": tickets,
	})
}

// GetMyTickets retrieves attendee's tickets
func (h *AttendeeHandler) GetMyTickets(c *gin.Context) {
	attendeeID, _ := middleware.GetUserID(c)

	var tickets []models.Ticket
	if err := h.db.Preload("Event").Preload("TicketType").Where("attendee_id = ?", attendeeID).Order("created_at DESC").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

// GetTicketDetails retrieves details of a specific ticket
func (h *AttendeeHandler) GetTicketDetails(c *gin.Context) {
	ticketID := c.Param("id")
	attendeeID, _ := middleware.GetUserID(c)

	var ticket models.Ticket
	if err := h.db.Preload("Event").Preload("TicketType").Preload("Transaction").First(&ticket, "id = ? AND attendee_id = ?", ticketID, attendeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// DownloadTicketPDF downloads the ticket PDF
func (h *AttendeeHandler) DownloadTicketPDF(c *gin.Context) {
	ticketID := c.Param("id")
	attendeeID, _ := middleware.GetUserID(c)

	var ticket models.Ticket
	if err := h.db.First(&ticket, "id = ? AND attendee_id = ?", ticketID, attendeeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	if ticket.PDFURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "PDF not available"})
		return
	}

	// Get PDF from storage
	pdfData, err := h.storageService.GetFile(ticket.PDFURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve PDF"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=ticket-%s.pdf", ticket.TicketNumber))
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// GetTransactionHistory retrieves attendee's transaction history
func (h *AttendeeHandler) GetTransactionHistory(c *gin.Context) {
	attendeeID, _ := middleware.GetUserID(c)

	var transactions []models.Transaction
	if err := h.db.Preload("Event").Where("user_id = ?", attendeeID).Order("created_at DESC").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
