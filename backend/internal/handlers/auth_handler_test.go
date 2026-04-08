package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/models"
)

func TestRegisterRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		payload RegisterRequest
		wantErr bool
	}{
		{
			name: "Valid registration",
			payload: RegisterRequest{
				Email:     "test@example.com",
				Password:  "Password123",
				FirstName: "Test",
				LastName:  "User",
				Role:      models.RoleAttendee,
			},
			wantErr: false,
		},
		{
			name: "Invalid email",
			payload: RegisterRequest{
				Email:     "invalid-email",
				Password:  "Password123",
				FirstName: "Test",
				LastName:  "User",
				Role:      models.RoleAttendee,
			},
			wantErr: true,
		},
		{
			name: "Short password",
			payload: RegisterRequest{
				Email:     "test@example.com",
				Password:  "short",
				FirstName: "Test",
				LastName:  "User",
				Role:      models.RoleAttendee,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonData, _ := json.Marshal(tt.payload)
			c.Request = httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
			c.Request.Header.Set("Content-Type", "application/json")

			var req RegisterRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantErr && err == nil {
				t.Error("Expected validation error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestLoginRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		payload LoginRequest
		wantErr bool
	}{
		{
			name: "Valid login",
			payload: LoginRequest{
				Email:    "test@example.com",
				Password: "Password123",
			},
			wantErr: false,
		},
		{
			name: "Missing email",
			payload: LoginRequest{
				Email:    "",
				Password: "Password123",
			},
			wantErr: true,
		},
		{
			name: "Missing password",
			payload: LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonData, _ := json.Marshal(tt.payload)
			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			c.Request.Header.Set("Content-Type", "application/json")

			var req LoginRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantErr && err == nil {
				t.Error("Expected validation error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestNewAuthHandler(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:      "test-secret",
		JWTExpiryHours: 24,
	}

	handler := NewAuthHandler(nil, cfg, nil, nil)

	if handler == nil {
		t.Error("AuthHandler should not be nil")
	}

	if handler.cfg != cfg {
		t.Error("Config not set correctly")
	}
}
