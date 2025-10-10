package response

import (
	"github.com/gin-gonic/gin"
)

// JSON sends a structured JSON response.
func JSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// Error sends a structured JSON error response.
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}
