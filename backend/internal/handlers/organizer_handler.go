package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/middleware"
	"github.com/warui/event-ticketing-api/internal/models"
	"github.com/warui/event-ticketing-api/internal/services"
	"gorm.io/gorm"
)

type OrganizerHandler struct {
	db             *gorm.DB
	cfg            *config.Config
	storageService *services.StorageService
	imageService   *services.ImageService
	metricsService *services.MetricsService
}

func NewOrganizerHandler(db *gorm.DB, cfg *config.Config, storageService *services.StorageService, imageService *services.ImageService) *OrganizerHandler {
	return &OrganizerHandler{
		db:             db,
		cfg:            cfg,
		storageService: storageService,
		imageService:   imageService,
		metricsService: services.NewMetricsService(db),
	}
}

type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Category    string    `json:"category" binding:"required"`
	Venue       string    `json:"venue" binding:"required"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	VideoURL    string    `json:"video_url"` // Optional promotional video URL
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
}

type CreateTicketTypeRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price" binding:"required,min=0"`
	Quantity    int       `json:"quantity" binding:"required,min=1"`
	MaxPerOrder int       `json:"max_per_order" binding:"required,min=1"`
	SaleStart   time.Time `json:"sale_start" binding:"required"`
	SaleEnd     time.Time `json:"sale_end" binding:"required"`
}

// CreateEvent creates a new event
func (h *OrganizerHandler) CreateEvent(c *gin.Context) {
	organizerID, _ := middleware.GetUserID(c)

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	event := &models.Event{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Venue:       req.Venue,
		Address:     req.Address,
		City:        req.City,
		Country:     req.Country,
		VideoURL:    req.VideoURL,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		OrganizerID: organizerID,
		Status:      models.EventStatusPending, // Auto-submit for approval
	}

	if err := h.db.Create(event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// UploadEventImage uploads an image for an event
func (h *OrganizerHandler) UploadEventImage(c *gin.Context) {
	eventID := c.Param("id")
	organizerID, _ := middleware.GetUserID(c)

	var event models.Event
	if err := h.db.First(&event, "id = ? AND organizer_id = ?", eventID, organizerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file required"})
		return
	}

	// Open file
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer fileContent.Close()

	// Read file content
	imageData := make([]byte, file.Size)
	if _, err := fileContent.Read(imageData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
		return
	}

	// Validate image
	if err := h.imageService.ValidateImage(imageData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image file"})
		return
	}

	// Process image
	processedImage, err := h.imageService.ProcessEventImage(imageData, 1200, 800)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		return
	}

	// Generate filename
	filename := services.GenerateUniqueFilename(fmt.Sprintf("event-%s", event.ID.String()[:8]), "jpg")

	// Upload to storage
	imageURL, err := h.storageService.UploadFile(processedImage, "events", filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Update event
	event.ImageURL = imageURL
	if err := h.db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}

// UpdateEvent updates an event
func (h *OrganizerHandler) UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")
	organizerID, _ := middleware.GetUserID(c)

	var event models.Event
	if err := h.db.First(&event, "id = ? AND organizer_id = ?", eventID, organizerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Only allow updates for draft or rejected events
	if event.Status != models.EventStatusDraft && event.Status != models.EventStatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update event in current status"})
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.Title = req.Title
	event.Description = req.Description
	event.Category = req.Category
	event.Venue = req.Venue
	event.Address = req.Address
	event.City = req.City
	event.Country = req.Country
	event.VideoURL = req.VideoURL
	event.StartDate = req.StartDate
	event.EndDate = req.EndDate

	if err := h.db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// SubmitEventForReview submits an event for moderation
func (h *OrganizerHandler) SubmitEventForReview(c *gin.Context) {
	eventID := c.Param("id")
	organizerID, _ := middleware.GetUserID(c)

	var event models.Event
	if err := h.db.Preload("TicketTypes").First(&event, "id = ? AND organizer_id = ?", eventID, organizerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.Status != models.EventStatusDraft && event.Status != models.EventStatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event cannot be submitted in current status"})
		return
	}

	// Validate event has at least one ticket type
	if len(event.TicketTypes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event must have at least one ticket type"})
		return
	}

	event.Status = models.EventStatusPending
	if err := h.db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event submitted for review", "event": event})
}

// PublishEvent publishes an approved event
func (h *OrganizerHandler) PublishEvent(c *gin.Context) {
	eventID := c.Param("id")
	organizerID, _ := middleware.GetUserID(c)

	var event models.Event
	if err := h.db.First(&event, "id = ? AND organizer_id = ?", eventID, organizerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.Status != models.EventStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only approved events can be published"})
		return
	}

	event.Status = models.EventStatusPublished
	if err := h.db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetMyEvents retrieves organizer's events
func (h *OrganizerHandler) GetMyEvents(c *gin.Context) {
	organizerID, _ := middleware.GetUserID(c)
	status := c.Query("status")

	query := h.db.Preload("TicketTypes").Where("organizer_id = ?", organizerID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var events []models.Event
	if err := query.Order("created_at DESC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetMyEvent retrieves a single event belonging to the organizer
func (h *OrganizerHandler) GetMyEvent(c *gin.Context) {
	organizerID, _ := middleware.GetUserID(c)
	eventID := c.Param("id")

	var event models.Event
	if err := h.db.Preload("TicketTypes").First(&event, "id = ? AND organizer_id = ?", eventID, organizerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// CreateTicketType creates a ticket type for an event
func (h *OrganizerHandler) CreateTicketType(c *gin.Context) {
	eventID := c.Param("id")
	organizerID, _ := middleware.GetUserID(c)

	var event models.Event
	if err := h.db.First(&event, "id = ? AND organizer_id = ?", eventID, organizerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var req CreateTicketTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate dates
	if req.SaleEnd.Before(req.SaleStart) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sale end date must be after sale start date"})
		return
	}

	eventIDUUID, _ := uuid.Parse(eventID)
	ticketType := &models.TicketType{
		EventID:     eventIDUUID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Quantity:    req.Quantity,
		MaxPerOrder: req.MaxPerOrder,
		SaleStart:   req.SaleStart,
		SaleEnd:     req.SaleEnd,
		IsActive:    true,
	}

	if err := h.db.Create(ticketType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket type"})
		return
	}

	c.JSON(http.StatusCreated, ticketType)
}

// GetOrganizerBalance retrieves organizer's balance
func (h *OrganizerHandler) GetOrganizerBalance(c *gin.Context) {
	organizerID, _ := middleware.GetUserID(c)

	var balance models.OrganizerBalance
	if err := h.db.Where("organizer_id = ?", organizerID).First(&balance).Error; err != nil {
		// Create balance if doesn't exist
		balance = models.OrganizerBalance{
			OrganizerID: organizerID,
		}
		h.db.Create(&balance)
	}

	c.JSON(http.StatusOK, balance)
}

// RequestWithdrawal creates a withdrawal request
func (h *OrganizerHandler) RequestWithdrawal(c *gin.Context) {
	organizerID, _ := middleware.GetUserID(c)

	var req struct {
		Amount        float64 `json:"amount" binding:"required,min=0"`
		BankName      string  `json:"bank_name" binding:"required"`
		AccountNumber string  `json:"account_number" binding:"required"`
		AccountName   string  `json:"account_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get platform settings
	var settings models.PlatformSettings
	if err := h.db.First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch platform settings"})
		return
	}

	// Check minimum withdrawal amount
	if req.Amount < settings.MinWithdrawalAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Minimum withdrawal amount is %.2f", settings.MinWithdrawalAmount),
		})
		return
	}

	// Get organizer balance
	var balance models.OrganizerBalance
	if err := h.db.Where("organizer_id = ?", organizerID).First(&balance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Balance not found"})
		return
	}

	// Check available balance
	if balance.AvailableBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Calculate withdrawal fee
	withdrawalFee := req.Amount * (settings.WithdrawalFeePercentage / 100)
	netAmount := req.Amount - withdrawalFee

	// Create withdrawal request
	withdrawal := &models.WithdrawalRequest{
		OrganizerID:   organizerID,
		Amount:        req.Amount,
		WithdrawalFee: withdrawalFee,
		NetAmount:     netAmount,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		Status:        models.WithdrawalStatusPending,
	}

	if err := h.db.Create(withdrawal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create withdrawal request"})
		return
	}

	// Update balance
	balance.AvailableBalance -= req.Amount
	balance.PendingBalance += req.Amount
	h.db.Save(&balance)

	c.JSON(http.StatusCreated, withdrawal)
}

// GetMyWithdrawals retrieves organizer's withdrawal requests
func (h *OrganizerHandler) GetMyWithdrawals(c *gin.Context) {
	organizerID, _ := middleware.GetUserID(c)

	var withdrawals []models.WithdrawalRequest
	if err := h.db.Where("organizer_id = ?", organizerID).Order("created_at DESC").Find(&withdrawals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawals"})
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}

// GetEventStats retrieves statistics for an event
func (h *OrganizerHandler) GetEventStats(c *gin.Context) {
	eventID := c.Param("id")
	organizerID, _ := middleware.GetUserID(c)

	var event models.Event
	if err := h.db.First(&event, "id = ? AND organizer_id = ?", eventID, organizerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var baseStats struct {
		TotalTicketsSold int64   `json:"total_tickets_sold"`
		TotalRevenue     float64 `json:"total_revenue"`
		NetRevenue       float64 `json:"net_revenue"`
		CheckedInTickets int64   `json:"checked_in_tickets"`
	}

	h.db.Model(&models.Ticket{}).Where("event_id = ? AND status = ?", eventID, models.TicketStatusConfirmed).Count(&baseStats.TotalTicketsSold)
	h.db.Model(&models.Ticket{}).Where("event_id = ? AND checked_in_at IS NOT NULL", eventID).Count(&baseStats.CheckedInTickets)

	var tickets []models.Ticket
	h.db.Where("event_id = ? AND status = ?", eventID, models.TicketStatusConfirmed).Find(&tickets)
	for _, ticket := range tickets {
		baseStats.TotalRevenue += ticket.Price
	}

	var settings models.PlatformSettings
	h.db.First(&settings)
	platformFee := baseStats.TotalRevenue * (settings.PlatformFeePercentage / 100)
	baseStats.NetRevenue = baseStats.TotalRevenue - platformFee

	// Extended metrics
	ticketTypeBreakdown, err := h.metricsService.GetEventTicketTypeBreakdown(eventUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ticket type breakdown"})
		return
	}

	dailyRevenue, err := h.metricsService.GetEventDailyRevenue(eventUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch daily revenue"})
		return
	}

	checkInStats, err := h.metricsService.GetEventCheckInStats(eventUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch check-in stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_tickets_sold":    baseStats.TotalTicketsSold,
		"total_revenue":         baseStats.TotalRevenue,
		"net_revenue":           baseStats.NetRevenue,
		"checked_in_tickets":    baseStats.CheckedInTickets,
		"ticket_type_breakdown": ticketTypeBreakdown,
		"daily_revenue":         dailyRevenue,
		"check_in_stats":        checkInStats,
	})
}

// VerifyTicket verifies and checks in a ticket
func (h *OrganizerHandler) VerifyTicket(c *gin.Context) {
	eventID := c.Param("id")
	organizerID, _ := middleware.GetUserID(c)

	var req struct {
		TicketNumber string `json:"ticket_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify event belongs to organizer
	var event models.Event
	if err := h.db.First(&event, "id = ? AND organizer_id = ?", eventID, organizerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Find ticket
	var ticket models.Ticket
	if err := h.db.Preload("TicketType").Preload("Attendee").First(&ticket, "ticket_number = ? AND event_id = ?", req.TicketNumber, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found or does not belong to this event"})
		return
	}

	// Check if already used
	if ticket.Status == models.TicketStatusUsed {
		c.JSON(http.StatusConflict, gin.H{
			"error":         "Ticket already used",
			"ticket":        ticket,
			"checked_in_at": ticket.CheckedInAt,
		})
		return
	}

	// Check if cancelled
	if ticket.Status == models.TicketStatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket is cancelled"})
		return
	}

	// Check if confirmed
	if ticket.Status != models.TicketStatusConfirmed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket is not valid (status: " + string(ticket.Status) + ")"})
		return
	}

	// Update ticket status
	now := time.Now()
	ticket.Status = models.TicketStatusUsed
	ticket.CheckedInAt = &now
	ticket.CheckedInBy = &organizerID

	if err := h.db.Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket verified and checked in successfully",
		"ticket":  ticket,
	})
}
