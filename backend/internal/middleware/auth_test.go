package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/auth"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/models"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret:      "test-secret",
		JWTExpiryHours: 24,
	}

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  models.RoleAttendee,
	}

	token, _ := auth.GenerateToken(user, cfg.JWTSecret, cfg.JWTExpiryHours)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			AuthMiddleware(cfg)(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       models.Role
		requiredRoles  []models.Role
		expectedStatus int
	}{
		{
			name:           "Admin accessing admin endpoint",
			userRole:       models.RoleAdmin,
			requiredRoles:  []models.Role{models.RoleAdmin},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Moderator accessing moderator endpoint",
			userRole:       models.RoleModerator,
			requiredRoles:  []models.Role{models.RoleModerator, models.RoleAdmin},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Attendee accessing admin endpoint",
			userRole:       models.RoleAttendee,
			requiredRoles:  []models.Role{models.RoleAdmin},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Organizer accessing moderator endpoint",
			userRole:       models.RoleOrganizer,
			requiredRoles:  []models.Role{models.RoleModerator, models.RoleAdmin},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			c.Set("user_role", tt.userRole)

			RequireRole(tt.requiredRoles...)(c)

			if w.Code != tt.expectedStatus && w.Code != 0 {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID)

	retrievedID, err := GetUserID(c)
	if err != nil {
		t.Fatalf("GetUserID failed: %v", err)
	}

	if retrievedID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, retrievedID)
	}
}

func TestGetUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	role := models.RoleOrganizer

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_role", role)

	retrievedRole, err := GetUserRole(c)
	if err != nil {
		t.Fatalf("GetUserRole failed: %v", err)
	}

	if retrievedRole != role {
		t.Errorf("Expected role %v, got %v", role, retrievedRole)
	}
}
