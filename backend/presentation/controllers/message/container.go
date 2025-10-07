package message

import (
	"backend/application/usecases"
	"backend/infrastructure/persistence"
	"database/sql"

	"github.com/go-playground/validator/v10"
)

type Container struct {
	Controller *MessageController
}

func New(db *sql.DB, validate *validator.Validate) *Container {
	messageRepository := persistence.NewPgMessageRepository(db)
	createMessageUseCase := usecases.NewCreateMessageUseCase(messageRepository)
	messageController := NewMessageController(createMessageUseCase, validate)

	return &Container{
		Controller: messageController,
	}
}
