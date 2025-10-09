package main

import (
	"context"
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
	"backend/internal/shared/infra/redis"
	services "backend/internal/shared/infra/services"
	"backend/internal/users"

	"github.com/go-playground/validator/v10"
)

func main() {
	log.Println("Starting modular server...")
	ctx := context.Background()

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

	redisClient, err := redis.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("could not initialize redis client: %s", err)
	}

	store := db.NewSQLStore(database)

	jwtService, err := services.NewJWTService(cfg.JWTSecret)
	if err != nil {
		log.Fatalf("Failed to create JWT service: %v", err)
	}

	mailerConfig := application.MailerConfig{
		SMTP_HOST:     cfg.SMTPHost,
		SMTP_PORT:     cfg.SMTPPort,
		SMTP_FROM:     cfg.SMTPFrom,
		SMTP_PASSWORD: cfg.SMTPPassword,
	}
	mailerService, err := services.NewMailerService(mailerConfig)
	if err != nil {
		log.Fatalf("Failed to create Mailer service: %v", err)
	}

	// --- Repositories ---
	authRepository := auth_persistence.NewPgAuthRepository(database)
	tokenRepository := auth_persistence.NewRedisTokenRepository(redisClient)
	messageRepository := message_persistence.NewPgMessageRepository(database)

	// --- Server Setup ---
	server := http.NewServer(cfg)
	apiV1 := server.Router.Group("/api/v1")

	// --- Module Registration ---
	// Each module is responsible for setting up its own dependencies and routes.
	auth.RegisterModule(authRepository, tokenRepository, apiV1, validate, jwtService, mailerService, store)
	users.RegisterModule(authRepository, apiV1, validate, mailerService, store, http.JWTMiddleware(jwtService))
	message.RegisterModule(messageRepository, apiV1, validate, http.JWTMiddleware(jwtService))

	// --- Start Server ---
	server.Run()
}
