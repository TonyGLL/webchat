package presentation

import (
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/http/request"
	"backend/internal/shared/http/response"
	"backend/internal/users/application/usecases"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UsersController struct {
	DeactivateUserUseCase *usecases.DeactivateUserUseCase
	GetUserProfileUseCase *usecases.GetUserProfileUseCase
	Validate              *validator.Validate
}

func NewUsersController(
	deactivateUserUseCase *usecases.DeactivateUserUseCase,
	getUserProfileUseCase *usecases.GetUserProfileUseCase,
	validate *validator.Validate,
) *UsersController {
	return &UsersController{
		DeactivateUserUseCase: deactivateUserUseCase,
		GetUserProfileUseCase: getUserProfileUseCase,
		Validate:              validate,
	}
}

func (ctrl *UsersController) GetUserProfile(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	userProfile, err := ctrl.GetUserProfileUseCase.Execute(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, shared_domain.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, shared_domain.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, userProfile)
}

func (ctrl *UsersController) DeactivateUser(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, shared_domain.ErrorResponse(errors.New("user ID not found in context")))
		return
	}

	err := ctrl.DeactivateUserUseCase.Execute(ctx.Request.Context(), userID.(int))
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, shared_domain.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, shared_domain.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}
