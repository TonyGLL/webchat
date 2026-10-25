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
			ctx.JSON(http.StatusNotFound, shared_domain.ErrorResponse(err))
		default:
			log.Printf("failed to deactivate account: %v", err)
			ctx.JSON(http.StatusInternalServerError, shared_domain.ErrorResponse(err))
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
	// The document is streamed row by row; headers go out with the first byte so a failure
	// before that (unknown user, DB down) can still answer with a proper JSON error.
	w := &exportWriter{ctx: ctx}
	if err := ctrl.ExportAccountUseCase.Execute(ctx.Request.Context(), userID, w); err != nil {
		if w.started {
			// Too late for a status code: the truncated body is invalid JSON, so the client notices.
			log.Printf("account export interrupted: %v", err)
			return
		}
		if errors.Is(err, shared_domain.ErrNotFound) {
			response.Error(ctx, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("failed to export account: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "Failed to export data")
	}
}

// exportWriter sets the download headers right before the first byte is written.
type exportWriter struct {
	ctx     *gin.Context
	started bool
}

func (w *exportWriter) Write(p []byte) (int, error) {
	if !w.started {
		w.started = true
		w.ctx.Header("Content-Disposition", `attachment; filename="webchat-my-data.json"`)
		w.ctx.Header("Content-Type", "application/json; charset=utf-8")
		w.ctx.Status(http.StatusOK)
	}
	return w.ctx.Writer.Write(p)
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
