package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		simulateError  bool
		expectedStatus int
	}{
		{
			name:           "No error",
			simulateError:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With error",
			simulateError:  true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.Use(ErrorHandler())
			router.GET("/test", func(c *gin.Context) {
				if tt.simulateError {
					c.Error(errors.New("test error"))
				}
				c.Status(http.StatusOK)
			})

			c.Request = httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, c.Request)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
