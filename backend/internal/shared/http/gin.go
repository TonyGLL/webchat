package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/shared/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	config *config.Config
	Router *gin.Engine
}

// NewServer creates a new HTTP server and sets up shared middleware.
func NewServer(config *config.Config) *Server {
	router := gin.Default()
	server := &Server{
		config: config,
		Router: router,
	}
	server.setupMiddleware()
	return server
}

func (s *Server) setupMiddleware() {
	s.Router.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
}

// Run starts the HTTP server and handles graceful shutdown.
func (s *Server) Run() {
	port := s.config.Port
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: s.Router,
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
