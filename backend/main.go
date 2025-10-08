package main

import (
	"backend/application/usecases"
	"backend/infrastructure/config"
	"backend/infrastructure/database"
	"backend/infrastructure/jwt"
	"backend/infrastructure/persistence"
	"backend/presentation/controllers/auth"
	"backend/presentation/controllers/message"
	"backend/presentation/http"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
)

func main() {
	log.Println("Starting server...")

	// --- Configuration ---
	// All configuration is loaded from a single source of truth.
	cfg, err := config.NewConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- Dependency Injection (Composition Root) ---

	// Infrastructure Layer
	db, err := database.NewDBPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	defer db.Close()

	validate := validator.New()
	jwtService, err := jwt.NewJWTService(cfg.JWTSecret)
	if err != nil {
		log.Fatalf("Failed to create JWT service: %v", err)
	}

	authRepository := persistence.NewPgAuthRepository(db)
	messageRepository := persistence.NewPgMessageRepository(db)

	// Application Layer (Use Cases)
	loginUseCase := usecases.NewLoginUseCase(authRepository)
	registerUseCase := usecases.NewRegisterUseCase(authRepository)
	createMessageUseCase := usecases.NewCreateMessageUseCase(messageRepository)

	// Presentation Layer (Controllers)
	authController := auth.NewAuthController(loginUseCase, registerUseCase, jwtService, validate)
	messageController := message.NewMessageController(createMessageUseCase, validate)

	// --- Server Initialization ---
	// The server receives the configuration it needs, including CORS settings.
	server := http.NewServer(authController, messageController, jwtService, cfg)
	server.Run()
}
