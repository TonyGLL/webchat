package presentation

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"backend/internal/auth/application/dtos"
	"backend/internal/auth/application/usecases"
	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/shared/http/request"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthController struct {
	LoginUseCase                *usecases.LoginUseCase
	RegisterUseCase             *usecases.RegisterUseCase
	RefreshTokenUseCase         *usecases.RefreshTokenUseCase
	SendVerifyEmailUseCase      *usecases.SendVerifyEmailUseCase
	VerifyEmailUseCase          *usecases.VerifyEmailUseCase
	LogoutUseCase               *usecases.LogoutUseCase
	RequestPasswordResetUseCase *usecases.RequestPasswordResetUseCase
	ResetPasswordUseCase        *usecases.ResetPasswordUseCase
	Validate                    *validator.Validate
	// SecureCookies marks the refresh cookie as HTTPS-only (set in production).
	SecureCookies bool
}

func NewAuthController(
	loginUseCase *usecases.LoginUseCase,
	registerUseCase *usecases.RegisterUseCase,
	refreshTokenUseCase *usecases.RefreshTokenUseCase,
	sendVerifyEmailUseCase *usecases.SendVerifyEmailUseCase,
	verifyEmailUseCase *usecases.VerifyEmailUseCase,
	logoutUseCase *usecases.LogoutUseCase,
	requestPasswordResetUseCase *usecases.RequestPasswordResetUseCase,
	resetPasswordUseCase *usecases.ResetPasswordUseCase,
	validate *validator.Validate,
	secureCookies bool,
) *AuthController {
	return &AuthController{
		LoginUseCase:                loginUseCase,
		RegisterUseCase:             registerUseCase,
		RefreshTokenUseCase:         refreshTokenUseCase,
		SendVerifyEmailUseCase:      sendVerifyEmailUseCase,
		VerifyEmailUseCase:          verifyEmailUseCase,
		LogoutUseCase:               logoutUseCase,
		RequestPasswordResetUseCase: requestPasswordResetUseCase,
		ResetPasswordUseCase:        resetPasswordUseCase,
		Validate:                    validate,
		SecureCookies:               secureCookies,
	}
}

const (
	refreshCookieName = "refresh_token"
	// The cookie is only sent to the auth routes, never to the rest of the API.
	refreshCookiePath = "/api/v1/auth"
)

// sendSession stores the refresh token in an httpOnly cookie (invisible to JavaScript, so an
// XSS cannot steal it) and returns only the short-lived access token in the body.
func (ctrl *AuthController) sendSession(ctx *gin.Context, status int, session *dtos.AuthResponseDTO) {
	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie(refreshCookieName, session.RefreshToken, int(usecases.RefreshTokenDuration.Seconds()), refreshCookiePath, "", ctrl.SecureCookies, true)

	body := *session
	body.RefreshToken = ""
	ctx.JSON(status, body)
}

func (ctrl *AuthController) clearRefreshCookie(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie(refreshCookieName, "", -1, refreshCookiePath, "", ctrl.SecureCookies, true)
}

func (ctrl *AuthController) Login(ctx *gin.Context) {
	var input dtos.LoginInputDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	input.IP = ctx.ClientIP()
	response, err := ctrl.LoginUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		var throttled *domain.ThrottledError
		if errors.As(err, &throttled) {
			ctx.Header("Retry-After", strconv.Itoa(int(throttled.RetryAfter.Seconds())+1))
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many failed attempts, try again shortly"})
			return
		}
		if errors.Is(err, shared_domain.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log in"})
		return
	}

	ctrl.sendSession(ctx, http.StatusOK, response)
}

func (ctrl *AuthController) RefreshToken(ctx *gin.Context) {
	// The cookie is the normal source; a JSON body is accepted for non-browser clients.
	token, _ := ctx.Cookie(refreshCookieName)
	if token == "" {
		var input dtos.RefreshTokenInputDTO
		_ = ctx.ShouldBindJSON(&input)
		token = input.RefreshToken
	}
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	response, err := ctrl.RefreshTokenUseCase.Execute(ctx.Request.Context(), dtos.RefreshTokenInputDTO{RefreshToken: token})
	if err != nil {
		if errors.Is(err, shared_domain.ErrInvalidToken) {
			ctrl.clearRefreshCookie(ctx)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}
		log.Printf("failed to refresh token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	ctrl.sendSession(ctx, http.StatusOK, response)
}

// Logout revokes the refresh token and clears its cookie. It always succeeds.
func (ctrl *AuthController) Logout(ctx *gin.Context) {
	if token, err := ctx.Cookie(refreshCookieName); err == nil && token != "" {
		if err := ctrl.LogoutUseCase.Execute(ctx.Request.Context(), token); err != nil {
			log.Printf("failed to revoke refresh token: %v", err)
		}
	}
	ctrl.clearRefreshCookie(ctx)
	ctx.Status(http.StatusNoContent)
}

func (ctrl *AuthController) Register(ctx *gin.Context) {
	var input dtos.RegisterInputDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	response, err := ctrl.RegisterUseCase.Execute(ctx.Request.Context(), input)
	if err != nil {
		if errors.Is(err, shared_domain.ErrConflict) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
			return
		}
		if errors.Is(err, domain.ErrInvalidUsername) || errors.Is(err, domain.ErrReservedUsername) || errors.Is(err, domain.ErrInvalidDisplayName) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrPasswordTooShort) || errors.Is(err, domain.ErrPasswordTooLong) ||
			errors.Is(err, domain.ErrPasswordCommon) || errors.Is(err, domain.ErrPasswordPersonal) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("failed to register user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

func (ctrl *AuthController) SendVerifyEmail(ctx *gin.Context) {
	var input dtos.SendVerifyEmailInputDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := ctx.ShouldBindUri(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := ctrl.Validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Only the account owner may trigger the mail (this route used to be an open mail-bomb and
	// account-existence oracle).
	if callerID, err := request.GetUserID(ctx); err != nil || callerID != input.ID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if err := ctrl.SendVerifyEmailUseCase.Execute(ctx.Request.Context(), input); err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Printf("failed to send verification email: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Verification email sent"})
}

func (ctrl *AuthController) VerifyEmail(ctx *gin.Context) {
	var input dtos.VerifyEmailInputDTO
	if err := ctx.ShouldBindUri(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := ctrl.Validate.Struct(input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := ctrl.VerifyEmailUseCase.Execute(ctx.Request.Context(), input); err != nil {
		if errors.Is(err, shared_domain.ErrInvalidToken) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "This link is invalid or has expired"})
			return
		}
		log.Printf("failed to verify email: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// ForgotPassword handles POST /auth/forgot-password. The answer is always the same, whether or
// not the address has an account, so it cannot be used to discover who is registered.
func (ctrl *AuthController) ForgotPassword(ctx *gin.Context) {
	var input dtos.ForgotPasswordInputDTO
	if err := ctx.ShouldBindJSON(&input); err != nil || ctrl.Validate.Struct(input) != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "A valid email is required"})
		return
	}
	if err := ctrl.RequestPasswordResetUseCase.Execute(ctx.Request.Context(), input.Email); err != nil {
		log.Printf("password reset request failed: %v", err)
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "If the address has an account, a reset link is on its way"})
}

// ResetPassword handles POST /auth/reset-password.
func (ctrl *AuthController) ResetPassword(ctx *gin.Context) {
	var input dtos.ResetPasswordInputDTO
	if err := ctx.ShouldBindJSON(&input); err != nil || ctrl.Validate.Struct(input) != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "token and a password of 8 to 72 characters are required"})
		return
	}
	err := ctrl.ResetPasswordUseCase.Execute(ctx.Request.Context(), input.Token, input.Password)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, gin.H{"message": "Password updated"})
	case errors.Is(err, shared_domain.ErrInvalidToken):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "This link is invalid or has expired"})
	case errors.Is(err, domain.ErrPasswordTooShort), errors.Is(err, domain.ErrPasswordTooLong),
		errors.Is(err, domain.ErrPasswordCommon), errors.Is(err, domain.ErrPasswordPersonal):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		log.Printf("failed to reset password: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
	}
}
