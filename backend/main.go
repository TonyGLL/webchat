package main

import (
	"backend/application/usecases"
	"backend/infrastructure/database"
	"backend/infrastructure/persistence"
	"backend/presentation/controllers"
	"backend/presentation/http"
	"log"
)

func main() {
	// Infrastructure
	dbPool, err := database.NewDBPool()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	defer dbPool.Close()

	messageRepository := persistence.NewPgMessageRepository(dbPool)

	// Application
	createMessageUseCase := usecases.NewCreateMessageUseCase(messageRepository)

	// Presentation
	messageController := controllers.NewMessageController(createMessageUseCase)

	// Start server
	http.InitServer(messageController)
}