package usecases

import (
	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	AccessTokenDuration  = time.Hour * 1
	RefreshTokenDuration = time.Hour * 24 * 7 // 7 days
)

// dummyHash is compared against when the user does not exist, so an unknown account costs about
// as much time as a wrong password and the response time does not reveal which accounts exist.
var (
	dummyHash     string
	dummyHashOnce sync.Once
)

type LoginUseCase struct {
	authRepo        domain.AuthRepository
	tokenRepo       domain.TokenRepository
	passwordService domain.PasswordService
	jwtService      shared_app.JwtService
}

func NewLoginUseCase(
	authRepo domain.AuthRepository,
	tokenRepo domain.TokenRepository,
	ps domain.PasswordService,
	jwtS shared_app.JwtService) *LoginUseCase {
	return &LoginUseCase{
		authRepo:        authRepo,
		tokenRepo:       tokenRepo,
		passwordService: ps,
		jwtService:      jwtS,
	}
}

// Execute handles user login. It returns the authenticated user or an error.
func (uc *LoginUseCase) Execute(ctx context.Context, input dtos.LoginInputDTO) (*dtos.AuthResponseDTO, error) {
	user, err := uc.authRepo.GetUserByEmailOrUsername(ctx, strings.TrimSpace(input.User))
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			uc.spendHashTime(input.Password)
			return nil, shared_domain.ErrInvalidCredentials // User not found
		}
		return nil, err
	}

	hashedPassword, err := uc.authRepo.ValidateUserPassword(ctx, user.ID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			uc.spendHashTime(input.Password)
			return nil, shared_domain.ErrInvalidCredentials // Password not found for user
		}
		return nil, err
	}

	if !uc.passwordService.CheckPasswordHash(input.Password, hashedPassword) {
		return nil, shared_domain.ErrInvalidCredentials // Passwords do not match
	}

	if err := uc.authRepo.SetLastAccess(ctx, user.ID); err != nil {
		// This error might not be critical for the login flow itself.
		// For now, we return it, but it could also just be logged.
		return nil, err
	}

	// Short-lived access token
	accessToken, err := uc.jwtService.GenerateToken(user.ID, AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	// Long-lived refresh token; its ID is stored in Redis so it can be revoked and used only once
	refreshTokenID := uuid.New().String()
	refreshToken, err := uc.jwtService.GenerateRefreshToken(user.ID, refreshTokenID, RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	if err := uc.tokenRepo.StoreToken(ctx, user.ID, refreshTokenID, "refresh_token", RefreshTokenDuration); err != nil {
		return nil, err
	}

	response := uc.buildLoginResponse(user, accessToken, refreshToken)

	return &response, nil
}

// spendHashTime runs one bcrypt comparison whose result is ignored.
func (uc *LoginUseCase) spendHashTime(password string) {
	dummyHashOnce.Do(func() {
		dummyHash, _ = uc.passwordService.HashPassword("not-a-real-password")
	})
	uc.passwordService.CheckPasswordHash(password, dummyHash)
}

func (uc *LoginUseCase) buildLoginResponse(user *domain.User, accessToken, refreshToken string) dtos.AuthResponseDTO {
	userResponse := dtos.UserResponseDTO{
		ID:        user.ID,
		Name:      user.Name,
		LastName:  user.LastName,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		AvatarURL: user.AvatarUrl,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	return dtos.AuthResponseDTO{
		User:         userResponse,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
}
