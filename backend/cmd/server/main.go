package main

import (
	"log"
	"os"

	"backend/internal/auth"
	auth_persistence "backend/internal/auth/persistence"
	"backend/internal/message"
	message_persistence "backend/internal/message/persistence"
	"backend/internal/shared/application"
	"backend/internal/shared/config"
	"backend/internal/shared/http"
	"backend/internal/shared/infra/db"
	services "backend/internal/shared/infra/services"

	"github.com/go-playground/validator/v10"
)

func main() {
	log.Println("Starting modular server...")

	// --- Configuration ---
	cfg, err := config.NewConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- Shared Dependencies ---
	validate := validator.New()

	database, err := db.NewDBPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	defer database.Close()

	store := db.NewSQLStore(database)

	jwtService, err := services.NewJWTService(cfg.JWTSecret)
	if err != nil {
		log.Fatalf("Failed to create JWT service: %v", err)
	}

	mailerConfig := application.MailerConfig{
		SMTP_HOST:     cfg.SMTPHost,
		SMTP_FROM:     cfg.SMTPFrom,
		SMTP_PASSWORD: cfg.SMTPPassword,
	}
	mailerService, err := services.NewMailerService(mailerConfig)
	if err != nil {
		log.Fatalf("Failed to create Mailer service: %v", err)
	}

	// --- Repositories ---
	authRepository := auth_persistence.NewPgAuthRepository(database)
	messageRepository := message_persistence.NewPgMessageRepository(database)

	// --- Server Setup ---
	server := http.NewServer(cfg)
	apiV1 := server.Router.Group("/api/v1")

	// --- Module Registration ---
	// Each module is responsible for setting up its own dependencies and routes.
	auth.RegisterModule(authRepository, apiV1, validate, jwtService, mailerService, store)
	message.RegisterModule(messageRepository, apiV1, validate, http.JWTMiddleware(jwtService))

	// --- Start Server ---
	server.Run()
}
