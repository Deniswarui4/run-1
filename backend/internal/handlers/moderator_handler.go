package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/middleware"
	"github.com/warui/event-ticketing-api/internal/models"
	"github.com/warui/event-ticketing-api/internal/services"
	"gorm.io/gorm"
)

type ModeratorHandler struct {
	db           *gorm.DB
	cfg          *config.Config
	emailService *services.EmailService
}

func NewModeratorHandler(db *gorm.DB, cfg *config.Config, emailService *services.EmailService) *ModeratorHandler {
	return &ModeratorHandler{
		db:           db,
		cfg:          cfg,
		emailService: emailService,
	}
}

// GetPendingEvents retrieves events pending moderation
func (h *ModeratorHandler) GetPendingEvents(c *gin.Context) {
	var events []models.Event
	if err := h.db.Preload("Organizer").Where("status = ?", models.EventStatusPending).Order("created_at ASC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetEventForReview retrieves a specific event for review
func (h *ModeratorHandler) GetEventForReview(c *gin.Context) {
	eventID := c.Param("id")

	var event models.Event
	if err := h.db.Preload("Organizer").Preload("TicketTypes").First(&event, "id = ?", eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// ReviewEvent approves or rejects an event
func (h *ModeratorHandler) ReviewEvent(c *gin.Context) {
	eventID := c.Param("id")
	moderatorID, _ := middleware.GetUserID(c)

	var req struct {
		Action  string `json:"action" binding:"required,oneof=approve reject"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var event models.Event
	if err := h.db.Preload("Organizer").First(&event, "id = ?", eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.Status != models.EventStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event is not pending review"})
		return
	}

	if req.Action == "approve" {
		event.Status = models.EventStatusPublished // Auto-publish on approval
		event.ModeratorID = &moderatorID
		event.ModerationComment = req.Comment
		now := time.Now()
		event.ModeratedAt = &now
	} else if req.Action == "reject" {
		event.Status = models.EventStatusRejected
		event.ModeratorID = &moderatorID
		event.ModerationComment = req.Comment
		now := time.Now()
		event.ModeratedAt = &now
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action. Must be 'approve' or 'reject'"})
		return
	}

	if err := h.db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	// Send email notification to organizer
	go h.emailService.SendEventApprovalEmail(&event, &event.Organizer, req.Action == "approve")

	c.JSON(http.StatusOK, event)
}

// GetModerationStats returns moderation statistics
func (h *ModeratorHandler) GetModerationStats(c *gin.Context) {
	moderatorID, _ := middleware.GetUserID(c)

	var stats struct {
		PendingEvents  int64 `json:"pending_events"`
		ApprovedEvents int64 `json:"approved_events"`
		RejectedEvents int64 `json:"rejected_events"`
		MyReviews      int64 `json:"my_reviews"`
	}

	h.db.Model(&models.Event{}).Where("status = ?", models.EventStatusPending).Count(&stats.PendingEvents)
	h.db.Model(&models.Event{}).Where("status = ?", models.EventStatusApproved).Count(&stats.ApprovedEvents)
	h.db.Model(&models.Event{}).Where("status = ?", models.EventStatusRejected).Count(&stats.RejectedEvents)
	h.db.Model(&models.Event{}).Where("moderator_id = ?", moderatorID).Count(&stats.MyReviews)

	c.JSON(http.StatusOK, stats)
}

// GetMyReviews returns events reviewed by the current moderator
func (h *ModeratorHandler) GetMyReviews(c *gin.Context) {
	moderatorID, _ := middleware.GetUserID(c)

	var events []models.Event
	if err := h.db.Preload("Organizer").Where("moderator_id = ?", moderatorID).Order("moderated_at DESC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, events)
}
