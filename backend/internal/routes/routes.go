package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/handlers"
	"github.com/warui/event-ticketing-api/internal/middleware"
	"github.com/warui/event-ticketing-api/internal/services"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Initialize services
	storageService, _ := services.NewStorageService(cfg)
	emailService := services.NewEmailService(cfg)
	twoFAService := services.NewTwoFAService(cfg)
	pesapalService := services.NewPesapalService(cfg)
	qrcodeService := services.NewQRCodeService()
	pdfService := services.NewPDFService()
	imageService := services.NewImageService()
	cartService := services.NewCartService(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg, emailService, twoFAService)
	adminHandler := handlers.NewAdminHandler(db, cfg, emailService)
	moderatorHandler := handlers.NewModeratorHandler(db, cfg, emailService)
	organizerHandler := handlers.NewOrganizerHandler(db, cfg, storageService, imageService)
	attendeeHandler := handlers.NewAttendeeHandler(db, cfg, pesapalService, storageService, qrcodeService, pdfService, emailService)
	cartHandler := handlers.NewCartHandler(db, cfg, cartService)
	guestHandler := handlers.NewGuestHandler(db, cfg, pesapalService, storageService, qrcodeService, pdfService, emailService)

	// Rate limiter
	rate := limiter.Rate{
		Period: cfg.RateLimitWindow,
		Limit:  int64(cfg.RateLimitRequests),
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.RateLimiter(rate))

	// Public routes
	public := v1.Group("/")
	{
		// Auth routes
		auth := public.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/verify-email", authHandler.VerifyEmail)
			auth.POST("/resend-verification", authHandler.ResendVerification)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Public event routes
		events := public.Group("/events")
		{
			events.GET("", attendeeHandler.GetPublishedEvents)
			events.GET("/featured", adminHandler.GetFeaturedEvents)
			events.GET("/:id", attendeeHandler.GetEventDetails)
		}

		// Public category routes
		public.GET("/categories", adminHandler.GetCategories)
		public.GET("/settings", adminHandler.GetPlatformSettings)

		// Payment verification and callback
		public.GET("/payments/verify", attendeeHandler.VerifyPayment)
		public.GET("/payments/callback", attendeeHandler.VerifyPayment) // Pesapal redirect endpoint

		// Guest checkout routes
		guest := public.Group("/guest")
		{
			guest.POST("/checkout", guestHandler.InitiateGuestCheckout)
			guest.GET("/tickets", guestHandler.LookupGuestTickets)
		}
	}

	// Protected routes (require authentication)
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// Profile routes
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)

		// 2FA routes
		twofa := protected.Group("/2fa")
		{
			twofa.POST("/setup", authHandler.Setup2FA)
			twofa.POST("/enable", authHandler.Enable2FA)
			twofa.POST("/disable", authHandler.Disable2FA)
		}

		// 2FA verification (public route for login flow)
		public.POST("/auth/verify-2fa", authHandler.Verify2FA)

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireAdmin())
		{
			// Platform settings
			admin.PUT("/settings", adminHandler.UpdatePlatformSettings)

			// Withdrawal management
			admin.GET("/withdrawals", adminHandler.GetWithdrawalRequests)
			admin.POST("/withdrawals/:id/review", adminHandler.ReviewWithdrawalRequest)
			admin.POST("/withdrawals/:id/process", adminHandler.ProcessWithdrawal)

			// User management
			admin.GET("/users", adminHandler.GetAllUsers)
			admin.PUT("/users/:id/role", adminHandler.ManageUserRole)
			admin.PUT("/users/:id/toggle-status", adminHandler.ToggleUserStatus)

			// Statistics
			admin.GET("/stats", adminHandler.GetPlatformStats)

			// Category management
			admin.POST("/categories", adminHandler.CreateCategory)
			admin.PUT("/categories/:id", adminHandler.UpdateCategory)
			admin.DELETE("/categories/:id", adminHandler.DeleteCategory)

			// Featured events management
			admin.PATCH("/events/:id/featured", adminHandler.ToggleEventFeatured)
		}

		// Moderator routes
		moderator := protected.Group("/moderator")
		moderator.Use(middleware.RequireModerator())
		{
			moderator.GET("/events/pending", moderatorHandler.GetPendingEvents)
			moderator.GET("/events/:id", moderatorHandler.GetEventForReview)
			moderator.POST("/events/:id/review", moderatorHandler.ReviewEvent)
			moderator.GET("/stats", moderatorHandler.GetModerationStats)
			moderator.GET("/reviews", moderatorHandler.GetMyReviews)
		}

		// Organizer routes
		organizer := protected.Group("/organizer")
		organizer.Use(middleware.RequireOrganizer())
		{
			// Event management
			organizer.POST("/events", organizerHandler.CreateEvent)
			organizer.GET("/events", organizerHandler.GetMyEvents)
			organizer.GET("/events/:id", organizerHandler.GetMyEvent)
			organizer.PUT("/events/:id", organizerHandler.UpdateEvent)
			organizer.POST("/events/:id/image", organizerHandler.UploadEventImage)
			organizer.POST("/events/:id/submit", organizerHandler.SubmitEventForReview)
			organizer.POST("/events/:id/publish", organizerHandler.PublishEvent)
			organizer.GET("/events/:id/stats", organizerHandler.GetEventStats)

			// Ticket type management
			organizer.POST("/events/:id/ticket-types", organizerHandler.CreateTicketType)

			// Ticket verification
			organizer.POST("/events/:id/tickets/verify", organizerHandler.VerifyTicket)

			// Financial management
			organizer.GET("/balance", organizerHandler.GetOrganizerBalance)
			organizer.POST("/withdrawals", organizerHandler.RequestWithdrawal)
			organizer.GET("/withdrawals", organizerHandler.GetMyWithdrawals)
		}

		// Attendee routes (all authenticated users can purchase tickets)
		tickets := protected.Group("/tickets")
		{
			tickets.POST("/purchase", attendeeHandler.InitiateTicketPurchase)
			tickets.GET("/my-tickets", attendeeHandler.GetMyTickets)
			tickets.GET("/:id", attendeeHandler.GetTicketDetails)
			tickets.GET("/:id/download", attendeeHandler.DownloadTicketPDF)
		}

		// Transaction routes
		protected.GET("/transactions", attendeeHandler.GetTransactionHistory)
	}

	// Serve static files for local storage
	if cfg.StorageType == "local" {
		router.Static("/storage", cfg.LocalStoragePath)
	}

	// Optional-auth cart routes
	optionalAuth := v1.Group("/")
	optionalAuth.Use(middleware.OptionalAuthMiddleware(cfg))
	{
		optionalAuth.GET("/cart", cartHandler.GetCart)
		optionalAuth.PUT("/cart", cartHandler.UpsertCart)
		optionalAuth.DELETE("/cart", cartHandler.DeleteCart)
	}
}
