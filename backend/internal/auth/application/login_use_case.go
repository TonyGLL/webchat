package application

import (
	"context"
	"errors"
	"time"

	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
)

type LoginUseCase struct {
	store           shared_app.Store
	passwordService domain.PasswordService
	jwtService      shared_app.JwtService
}

func NewLoginUseCase(
	store shared_app.Store,
	ps domain.PasswordService,
	jwtS shared_app.JwtService) *LoginUseCase {
	return &LoginUseCase{
		store:           store,
		passwordService: ps,
		jwtService:      jwtS,
	}
}

// Execute handles user login. It returns the authenticated user or an error.
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInputDTO) (*AuthResponseDTO, error) {
	authRepo := uc.store.AuthRepository()

	user, err := authRepo.GetUserByEmailOrUsername(ctx, input.User)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials // User not found
		}
		return nil, err
	}

	hashedPassword, err := authRepo.ValidateUserPassword(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials // Password not found for user
		}
		return nil, err
	}

	if !uc.passwordService.CheckPasswordHash(input.Password, hashedPassword) {
		return nil, domain.ErrInvalidCredentials // Passwords do not match
	}

	if err := authRepo.SetLastAccess(ctx, user.ID); err != nil {
		// This error might not be critical for the login flow itself.
		// For now, we return it, but it could also just be logged.
		return nil, err
	}

	token, err := uc.jwtService.GenerateToken(user.ID, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	response := uc.buildLoginResponse(user, token)

	return &response, nil
}

func (uc *LoginUseCase) buildLoginResponse(user *domain.User, token string) AuthResponseDTO {
	userResponse := UserResponseDTO{
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
	return AuthResponseDTO{
		User:  userResponse,
		Token: token,
	}
}
