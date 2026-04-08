package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "GET request with CORS headers",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			checkHeaders:   true,
		},
		{
			name:           "POST request with CORS headers",
			method:         "POST",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.Use(CORS())
			router.Any("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			c.Request = httptest.NewRequest(tt.method, "/test", nil)
			router.ServeHTTP(w, c.Request)

			if tt.checkHeaders {
				if w.Header().Get("Access-Control-Allow-Origin") != "*" {
					t.Error("CORS header Access-Control-Allow-Origin not set correctly")
				}

				if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
					t.Error("CORS header Access-Control-Allow-Credentials not set correctly")
				}
			}
		})
	}
}
