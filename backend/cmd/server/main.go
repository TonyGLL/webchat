package main

import (
	"log"
	"os"

	"backend/internal/auth"
	"backend/internal/message"
	"backend/internal/shared/config"
	"backend/internal/shared/infra/db"
	"backend/internal/shared/infra/jwt"
	"backend/internal/shared/http"

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

	jwtService, err := jwt.NewJWTService(cfg.JWTSecret)
	if err != nil {
		log.Fatalf("Failed to create JWT service: %v", err)
	}

	// --- Server Setup ---
	server := http.NewServer(cfg)
	apiV1 := server.Router.Group("/api/v1")

	// --- Module Registration ---
	// Each module is responsible for setting up its own dependencies and routes.
	auth.RegisterModule(store, apiV1, validate, jwtService)
	message.RegisterModule(store, apiV1, validate, http.JWTMiddleware(jwtService))

	// --- Start Server ---
	server.Run()
}