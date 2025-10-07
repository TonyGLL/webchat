
package auth

import (
	"backend/application/usecases"
	"backend/infrastructure/persistence"
	"database/sql"

	"github.com/go-playground/validator/v10"
)

type Container struct {
	Controller *AuthController
}

func New(db *sql.DB, validate *validator.Validate) *Container {
	authRepository := persistence.NewPgAuthRepository(db)
	loginUseCase := usecases.NewLoginUseCase(authRepository)
	registerUseCase := usecases.NewRegisterUseCase(authRepository)
	authController := NewAuthController(loginUseCase, registerUseCase, validate)

	return &Container{
		Controller: authController,
	}
}
