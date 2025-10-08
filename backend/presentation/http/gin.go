package http

import (
	"backend/application/services"
	"backend/infrastructure/config"
	"backend/presentation/controllers/auth"
	"backend/presentation/controllers/message"
	"backend/presentation/http/middleware"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	authController    *auth.AuthController
	messageController *message.MessageController
	jwtService        services.JwtService
	config            *config.Config
}

// NewServer creates a new HTTP server with its dependencies.
func NewServer(
	authController *auth.AuthController,
	messageController *message.MessageController,
	jwtService services.JwtService,
	config *config.Config,
) *Server {
	return &Server{
		authController:    authController,
		messageController: messageController,
		jwtService:        jwtService,
		config:            config,
	}
}

// Run sets up the router and starts the server.
func (s *Server) Run() {
	router := gin.Default()

	// Middlewares
	router.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "OK"})
		})

		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/login", s.authController.Login)
			authRoutes.POST("/register", s.authController.Register)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTMiddleware(s.jwtService))
		{
			messageRoutes := protected.Group("/messages")
			{
				messageRoutes.POST("/", s.messageController.CreateMessage)
			}
		}
	}

	// Server setup and graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Server running on port: %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

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
