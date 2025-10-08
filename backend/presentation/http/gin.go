package http

import (
	jwt "backend/infrastructure/jwt"
	"backend/presentation/controllers/auth"
	"backend/presentation/controllers/message"
	"backend/presentation/http/middleware"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func InitServer(db *sql.DB, validate *validator.Validate) {
	router := gin.Default()

	// Middlewares
	/* CORS */
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://foo.com"},
		AllowMethods:     []string{"PUT", "PATCH", "OPTIONS", "GET", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
	}))

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	authContainer := auth.New(db, validate)
	messageContainer := message.New(db, validate)

	// Auth Routes
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authContainer.Controller.Login)
			auth.POST("/register", authContainer.Controller.Register)
		}
	}

	jwtImpl := jwt.NewJWTService()
	v1.Use(middleware.JWTMiddleware(jwtImpl))
	// Message routes
	{
		messages := v1.Group("/messages")
		{
			messages.POST("/", messageContainer.Controller.CreateMessage)
		}
	}

	// Get port from environment variables, with a default
	port := os.Getenv("PORT")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Server running on port: %s", port)
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
