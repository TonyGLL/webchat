package presentation

import (
	"backend/internal/users/application/usecases"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UsersController struct {
	DeactivateUserUseCase *usecases.DeactivateUserUseCase
	Validate              *validator.Validate
}

func NewUsersController(
	deactivateUserUseCase *usecases.DeactivateUserUseCase,
	validate *validator.Validate,
) *UsersController {
	return &UsersController{
		DeactivateUserUseCase: deactivateUserUseCase,
		Validate:              validate,
	}
}

func (ctrl *UsersController) DeactivateUser(ctx *gin.Context) {

}
