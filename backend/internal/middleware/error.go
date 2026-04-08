package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Log the error
			log.Printf("Error: %v", err.Err)

			// Return error response
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "An internal error occurred",
				"message": err.Error(),
			})
		}
	}
}
