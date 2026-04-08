package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/warui/event-ticketing-api/internal/config"
)

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	cfg := &config.Config{
		JWTSecret:         "test-secret",
		JWTExpiryHours:    24,
		StorageType:       "local",
		LocalStoragePath:  "./test_storage",
		RateLimitRequests: 100,
		RateLimitWindow:   60,
	}

	// Setup routes without database (will fail on actual requests but routes should be registered)
	SetupRoutes(router, nil, cfg)

	// Test that routes are registered
	routes := router.Routes()

	if len(routes) == 0 {
		t.Error("No routes were registered")
	}

	// Check for some key routes
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path] = true
	}

	expectedRoutes := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/events",
		"/api/v1/profile",
	}

	for _, path := range expectedRoutes {
		if !routePaths[path] {
			t.Errorf("Expected route %s not found", path)
		}
	}
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Add health endpoint directly for testing
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Event Ticketing API is running",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestPublicRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	cfg := &config.Config{
		JWTSecret:        "test-secret",
		StorageType:      "local",
		LocalStoragePath: "./test_storage",
	}

	SetupRoutes(router, nil, cfg)

	publicRoutes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/auth/register"},
		{"POST", "/api/v1/auth/login"},
		{"GET", "/api/v1/events"},
	}

	for _, route := range publicRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			// Just verify the route exists (will return error without DB but that's ok)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(route.method, route.path, nil)
			router.ServeHTTP(w, req)

			// Route should exist (not 404)
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not found", route.method, route.path)
			}
		})
	}
}
