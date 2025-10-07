package main

import (
	"backend/infrastructure/config"
	"backend/infrastructure/database"
	"backend/presentation/http"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
)

func main() {
	// Load environment variables from the file specified by CONFIG_FILE.
	// This is the approach used in the reference project.
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		// In a container, the file is copied to .env.
		// When running locally, the Makefile sets CONFIG_FILE.
		// This is a fallback for running directly without make (e.g. in an IDE).
		configFile = ".env"
	}
	err := config.LoadConfig(configFile)
	if err != nil {
		// This is not treated as a fatal error because environment variables
		// can be provided by other means (e.g., docker-compose, system env).
		log.Printf("Warning: could not load config from file '%s'. Relying on environment variables. Error: %v", configFile, err)
	}

	// Infrastructure
	db, err := database.NewDBPool()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	defer db.Close()

	// Presentation
	validate := validator.New()

	// Start server
	http.InitServer(db, validate)
}
