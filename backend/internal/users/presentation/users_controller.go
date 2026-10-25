package presentation

import (
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/http/request"
	"backend/internal/shared/http/response"
	"backend/internal/users/application/usecases"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UsersController struct {
	DeactivateUserUseCase *usecases.DeactivateUserUseCase
	GetUserProfileUseCase *usecases.GetUserProfileUseCase
	ExportAccountUseCase  *usecases.ExportAccountUseCase
	DeleteAccountUseCase  *usecases.DeleteAccountUseCase
	ListDirectoryUseCase  *usecases.ListDirectoryUseCase
	Validate              *validator.Validate
}

func NewUsersController(
	deactivateUserUseCase *usecases.DeactivateUserUseCase,
	getUserProfileUseCase *usecases.GetUserProfileUseCase,
	exportAccountUseCase *usecases.ExportAccountUseCase,
	deleteAccountUseCase *usecases.DeleteAccountUseCase,
	listDirectoryUseCase *usecases.ListDirectoryUseCase,
	validate *validator.Validate,
) *UsersController {
	return &UsersController{
		DeactivateUserUseCase: deactivateUserUseCase,
		GetUserProfileUseCase: getUserProfileUseCase,
		ExportAccountUseCase:  exportAccountUseCase,
		DeleteAccountUseCase:  deleteAccountUseCase,
		ListDirectoryUseCase:  listDirectoryUseCase,
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
			response.Error(ctx, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("failed to load profile: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to load profile")
		return
	}

	ctx.JSON(http.StatusOK, userProfile)
}

func (ctrl *UsersController) DeactivateUser(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body deleteAccountBody // same shape: the current password
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.Error(ctx, http.StatusBadRequest, "password is required")
		return
	}

	err := ctrl.DeactivateUserUseCase.Execute(ctx.Request.Context(), userID.(int), body.Password)
	if err != nil {
		switch {
		case errors.Is(err, shared_domain.ErrInvalidCredentials):
			response.Error(ctx, http.StatusForbidden, "Wrong password")
		case errors.Is(err, shared_domain.ErrNotFound):
			response.Error(ctx, http.StatusNotFound, "User not found")
		default:
			log.Printf("failed to deactivate account: %v", err)
			response.Error(ctx, http.StatusInternalServerError, "Failed to deactivate account")
		}
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

// ListDirectory handles the endpoint for GET /api/v1/users/directory?q=&limit=
func (ctrl *UsersController) ListDirectory(ctx *gin.Context) {
	userID, err := request.GetUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	limit, _ := strconv.Atoi(ctx.Query("limit"))

	entries, err := ctrl.ListDirectoryUseCase.Execute(ctx.Request.Context(), userID, ctx.Query("q"), limit)
	if err != nil {
		log.Printf("list directory: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "internal server error")
		return
	}
	response.JSON(ctx, http.StatusOK, entries)
}
