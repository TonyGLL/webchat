package http

import (
	"backend/presentation/controllers"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"

	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func InitServer(messageController *controllers.MessageController) {
	router := gin.Default()

	// Middlewares
	router.Use(cors.Default())
	router.Use(secure.New(secure.DefaultConfig()))

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

	// Get port from environment variables, with a default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Server running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
