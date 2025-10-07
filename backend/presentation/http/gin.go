package http

import (
	"backend/presentation/controllers"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func InitServer(messageController *controllers.MessageController) {
	router := gin.Default()

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	// Message routes
	{
		messages := v1.Group("/messages")
		{
			messages.POST("/", messageController.CreateMessage)
		}
	}

	fmt.Println("Server running on port 3001")
	if err := router.Run(":3001"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
