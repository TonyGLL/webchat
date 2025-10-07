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
	{
		messages := v1.Group("/messages")
		{
			messages.POST("/", messageController.CreateMessage)
		}
	}

	fmt.Println("Server is running on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}