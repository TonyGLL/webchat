package presentation

import (
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/http/request"
	"backend/internal/shared/http/response"
	"backend/internal/users/application/usecases"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UsersController struct {
	DeactivateUserUseCase *usecases.DeactivateUserUseCase
	GetUserProfileUseCase *usecases.GetUserProfileUseCase
	ExportAccountUseCase  *usecases.ExportAccountUseCase
	DeleteAccountUseCase  *usecases.DeleteAccountUseCase
	Validate              *validator.Validate
}

func NewUsersController(
	deactivateUserUseCase *usecases.DeactivateUserUseCase,
	getUserProfileUseCase *usecases.GetUserProfileUseCase,
	exportAccountUseCase *usecases.ExportAccountUseCase,
	deleteAccountUseCase *usecases.DeleteAccountUseCase,
	validate *validator.Validate,
) *UsersController {
	return &UsersController{
		DeactivateUserUseCase: deactivateUserUseCase,
		GetUserProfileUseCase: getUserProfileUseCase,
		ExportAccountUseCase:  exportAccountUseCase,
		DeleteAccountUseCase:  deleteAccountUseCase,
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

// ExportAccount handles GET /users/export: a JSON file with everything stored about the user.
func (ctrl *UsersController) ExportAccount(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	export, err := ctrl.ExportAccountUseCase.Execute(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			response.Error(ctx, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("failed to export account: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to export data")
		return
	}
	ctx.Header("Content-Disposition", `attachment; filename="webchat-my-data.json"`)
	ctx.JSON(http.StatusOK, export)
}

type deleteAccountBody struct {
	Password string `json:"password" binding:"required"`
}

// DeleteAccount handles DELETE /users/me: permanent deletion, confirmed with the password.
func (ctrl *UsersController) DeleteAccount(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	var body deleteAccountBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.Error(ctx, http.StatusBadRequest, "password is required")
		return
	}
	if err := ctrl.DeleteAccountUseCase.Execute(ctx.Request.Context(), userID, body.Password); err != nil {
		switch {
		case errors.Is(err, shared_domain.ErrInvalidCredentials):
			response.Error(ctx, http.StatusForbidden, "Wrong password")
		case errors.Is(err, shared_domain.ErrNotFound):
			response.Error(ctx, http.StatusNotFound, "User not found")
		default:
			log.Printf("failed to delete account: %v", err)
			response.Error(ctx, http.StatusInternalServerError, "Failed to delete account")
		}
		return
	}
	ctx.Status(http.StatusNoContent)
}
