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

type AdminHandler struct {
	db             *gorm.DB
	cfg            *config.Config
	emailService   *services.EmailService
	metricsService *services.MetricsService
	featureService *services.FeatureService
}

func NewAdminHandler(db *gorm.DB, cfg *config.Config, emailService *services.EmailService) *AdminHandler {
	return &AdminHandler{
		db:             db,
		cfg:            cfg,
		emailService:   emailService,
		metricsService: services.NewMetricsService(db),
		featureService: services.NewFeatureService(db),
	}
}

// GetPlatformSettings retrieves current platform settings
func (h *AdminHandler) GetPlatformSettings(c *gin.Context) {
	var settings models.PlatformSettings
	if err := h.db.First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Platform settings not found"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdatePlatformSettings updates platform settings
func (h *AdminHandler) UpdatePlatformSettings(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		PlatformFeePercentage   *float64 `json:"platform_fee_percentage"`
		WithdrawalFeePercentage *float64 `json:"withdrawal_fee_percentage"`
		MinWithdrawalAmount     *float64 `json:"min_withdrawal_amount"`
		Currency                *string  `json:"currency"`
		AutoFeatureThreshold    *int     `json:"auto_feature_threshold"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var settings models.PlatformSettings
	if err := h.db.First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Platform settings not found"})
		return
	}

	if req.PlatformFeePercentage != nil {
		if *req.PlatformFeePercentage < 0 || *req.PlatformFeePercentage > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Platform fee must be between 0 and 100"})
			return
		}
		settings.PlatformFeePercentage = *req.PlatformFeePercentage
	}

	if req.WithdrawalFeePercentage != nil {
		if *req.WithdrawalFeePercentage < 0 || *req.WithdrawalFeePercentage > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal fee must be between 0 and 100"})
			return
		}
		settings.WithdrawalFeePercentage = *req.WithdrawalFeePercentage
	}

	if req.MinWithdrawalAmount != nil {
		if *req.MinWithdrawalAmount < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum withdrawal amount must be positive"})
			return
		}
		settings.MinWithdrawalAmount = *req.MinWithdrawalAmount
	}

	if req.Currency != nil {
		settings.Currency = *req.Currency
	}

	if req.AutoFeatureThreshold != nil {
		if *req.AutoFeatureThreshold < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Auto feature threshold must be non-negative"})
			return
		}
		settings.AutoFeatureThreshold = *req.AutoFeatureThreshold
	}

	settings.UpdatedBy = userID

	if err := h.db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetWithdrawalRequests retrieves all withdrawal requests
func (h *AdminHandler) GetWithdrawalRequests(c *gin.Context) {
	status := c.Query("status")

	query := h.db.Preload("Organizer").Preload("Reviewer")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var requests []models.WithdrawalRequest
	if err := query.Order("created_at DESC").Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawal requests"})
		return
	}

	c.JSON(http.StatusOK, requests)
}

// ReviewWithdrawalRequest approves or rejects a withdrawal request
func (h *AdminHandler) ReviewWithdrawalRequest(c *gin.Context) {
	requestID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	var req struct {
		Action  string `json:"action" binding:"required,oneof=approve reject"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var withdrawal models.WithdrawalRequest
	if err := h.db.Preload("Organizer").First(&withdrawal, "id = ?", requestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal request not found"})
		return
	}

	if withdrawal.Status != models.WithdrawalStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal request already reviewed"})
		return
	}

	now := time.Now()
	withdrawal.ReviewedBy = &userID
	withdrawal.ReviewedAt = &now
	withdrawal.ReviewComment = req.Comment

	if req.Action == "approve" {
		withdrawal.Status = models.WithdrawalStatusApproved
	} else {
		withdrawal.Status = models.WithdrawalStatusRejected

		// Return amount to organizer's available balance
		var balance models.OrganizerBalance
		if err := h.db.Where("organizer_id = ?", withdrawal.OrganizerID).First(&balance).Error; err == nil {
			balance.AvailableBalance += withdrawal.Amount
			balance.PendingBalance -= withdrawal.Amount
			h.db.Save(&balance)
		}
	}

	if err := h.db.Save(&withdrawal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update withdrawal request"})
		return
	}

	// Send email notification
	go h.emailService.SendWithdrawalStatusEmail(&withdrawal, &withdrawal.Organizer)

	c.JSON(http.StatusOK, withdrawal)
}

// ProcessWithdrawal marks a withdrawal as processed
func (h *AdminHandler) ProcessWithdrawal(c *gin.Context) {
	requestID := c.Param("id")

	var req struct {
		TransactionRef string `json:"transaction_ref" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var withdrawal models.WithdrawalRequest
	if err := h.db.Preload("Organizer").First(&withdrawal, "id = ?", requestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal request not found"})
		return
	}

	if withdrawal.Status != models.WithdrawalStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal must be approved before processing"})
		return
	}

	now := time.Now()
	withdrawal.Status = models.WithdrawalStatusProcessed
	withdrawal.ProcessedAt = &now
	withdrawal.TransactionRef = req.TransactionRef

	if err := h.db.Save(&withdrawal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process withdrawal"})
		return
	}

	// Update organizer balance
	var balance models.OrganizerBalance
	if err := h.db.Where("organizer_id = ?", withdrawal.OrganizerID).First(&balance).Error; err == nil {
		balance.WithdrawnAmount += withdrawal.NetAmount
		balance.PendingBalance -= withdrawal.Amount
		h.db.Save(&balance)
	}

	// Send email notification
	go h.emailService.SendWithdrawalStatusEmail(&withdrawal, &withdrawal.Organizer)

	c.JSON(http.StatusOK, withdrawal)
}

// GetPlatformStats returns platform statistics
func (h *AdminHandler) GetPlatformStats(c *gin.Context) {
	startDateStr := c.Query("start")
	endDateStr := c.Query("end")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		endDate = time.Now()
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end date must be after start date"})
		return
	}

	var stats struct {
		TotalUsers       int64   `json:"total_users"`
		TotalOrganizers  int64   `json:"total_organizers"`
		TotalEvents      int64   `json:"total_events"`
		TotalTicketsSold int64   `json:"total_tickets_sold"`
		TotalRevenue     float64 `json:"total_revenue"`
		PlatformRevenue  float64 `json:"platform_revenue"`
	}

	h.db.Model(&models.User{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&stats.TotalUsers)
	h.db.Model(&models.User{}).Where("role = ?", models.RoleOrganizer).Count(&stats.TotalOrganizers)
	h.db.Model(&models.Event{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.EventStatusPublished, startDate, endDate).Count(&stats.TotalEvents)
	h.db.Model(&models.Ticket{}).Where("status = ? AND created_at BETWEEN ? AND ?", models.TicketStatusConfirmed, startDate, endDate).Count(&stats.TotalTicketsSold)

	var transactions []models.Transaction
	h.db.Where("status = ? AND type = ? AND created_at BETWEEN ? AND ?", models.TransactionStatusCompleted, models.TransactionTypeTicketPurchase, startDate, endDate).Find(&transactions)
	for _, t := range transactions {
		stats.TotalRevenue += t.Amount
		stats.PlatformRevenue += t.PlatformFee
	}

	// Extended metrics
	ticketTypeBreakdown, err := h.metricsService.GetPlatformTicketTypeBreakdown(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ticket type breakdown"})
		return
	}

	dailyUserRegistrations, err := h.metricsService.GetDailyUserRegistrations(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch daily user registrations"})
		return
	}

	topEvents, err := h.metricsService.GetTopEventsByRevenue(startDate, endDate, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch top events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_users":              stats.TotalUsers,
		"total_organizers":         stats.TotalOrganizers,
		"total_events":             stats.TotalEvents,
		"total_tickets_sold":       stats.TotalTicketsSold,
		"total_revenue":            stats.TotalRevenue,
		"platform_revenue":         stats.PlatformRevenue,
		"ticket_type_breakdown":    ticketTypeBreakdown,
		"daily_user_registrations": dailyUserRegistrations,
		"top_events":               topEvents,
	})
}

// ManageUserRole updates a user's role
func (h *AdminHandler) ManageUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Role models.Role `json:"role" binding:"required,oneof=admin moderator organizer attendee"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	oldRole := user.Role
	user.Role = req.Role

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	// If promoted to organizer, create balance record
	if req.Role == models.RoleOrganizer && oldRole != models.RoleOrganizer {
		var balance models.OrganizerBalance
		if err := h.db.Where("organizer_id = ?", user.ID).First(&balance).Error; err != nil {
			balance = models.OrganizerBalance{
				OrganizerID: user.ID,
			}
			h.db.Create(&balance)
		}
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// ToggleUserStatus activates or deactivates a user
func (h *AdminHandler) ToggleUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.IsActive = !user.IsActive

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// GetAllUsers retrieves all users (admin only)
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	role := c.Query("role")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	var users []models.User
	query := h.db.Model(&models.User{})

	if role != "" {
		query = query.Where("role = ?", role)
	}

	var total int64
	query.Count(&total)

	offset := 0
	if page != "1" {
		var pageNum int
		if _, err := fmt.Sscanf(page, "%d", &pageNum); err == nil {
			var limitNum int
			fmt.Sscanf(limit, "%d", &limitNum)
			offset = (pageNum - 1) * limitNum
		}
	}

	if err := query.Offset(offset).Limit(20).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Remove passwords
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// ==================== Category Management ====================

// GetCategories retrieves all categories
func (h *AdminHandler) GetCategories(c *gin.Context) {
	var categories []models.Category
	if err := h.db.Order("name ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// CreateCategory creates a new category
func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if category with same name already exists
	var existingCategory models.Category
	if err := h.db.Where("name = ?", req.Name).First(&existingCategory).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Category with this name already exists"})
		return
	}

	category := models.Category{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		IsActive:    true,
	}

	if err := h.db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory updates an existing category
func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	categoryID := c.Param("id")

	var category models.Category
	if err := h.db.Where("id = ?", categoryID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
		Icon        *string `json:"icon"`
		IsActive    *bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		// Check if new name already exists for another category
		var existingCategory models.Category
		if err := h.db.Where("name = ? AND id != ?", *req.Name, categoryID).First(&existingCategory).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Category with this name already exists"})
			return
		}
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.Color != nil {
		category.Color = *req.Color
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := h.db.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory deletes a category
func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")

	var category models.Category
	if err := h.db.Where("id = ?", categoryID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Check if any events are using this category
	var eventCount int64
	if err := h.db.Model(&models.Event{}).Where("category = ?", category.Name).Count(&eventCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check category usage"})
		return
	}

	if eventCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Cannot delete category. %d events are using this category", eventCount)})
		return
	}

	if err := h.db.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// ==================== Featured Events Management ====================

// ToggleEventFeatured toggles the featured status of a published event.
func (h *AdminHandler) ToggleEventFeatured(c *gin.Context) {
	eventID := c.Param("id")

	id, err := uuid.Parse(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	event, err := h.featureService.ToggleFeatured(id, true)
	if err != nil {
		switch err {
		case services.ErrEventNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case services.ErrEventNotPublished:
			c.JSON(http.StatusBadRequest, gin.H{"error": "only published events can be featured"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event featured status"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            event.ID,
		"is_featured":   event.IsFeatured,
		"featured_type": event.FeaturedType,
		"status":        event.Status,
		"title":         event.Title,
	})
}

// GetFeaturedEvents retrieves all featured events
func (h *AdminHandler) GetFeaturedEvents(c *gin.Context) {
	var events []models.Event
	if err := h.db.Where("is_featured = ? AND status = ?", true, models.EventStatusPublished).
		Preload("Organizer").
		Preload("TicketTypes").
		Order("start_date ASC").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch featured events"})
		return
	}

	c.JSON(http.StatusOK, events)
}
