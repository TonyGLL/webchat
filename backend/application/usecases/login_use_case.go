package usecases

import (
	"backend/application"
	"backend/application/dtos"
	"backend/application/services"
	"backend/domain"
	presentation_dtos "backend/presentation/dtos"
	"context"
	"errors"
	"time"
)

type LoginUseCase struct {
	store           application.Store
	passwordService domain.PasswordService
	jwtService      services.JwtService
}

func NewLoginUseCase(
	store application.Store,
	ps domain.PasswordService,
	jwtS services.JwtService) *LoginUseCase {
	return &LoginUseCase{
		store:           store,
		passwordService: ps,
		jwtService:      jwtS,
	}
}

// Execute handles user login. It returns the authenticated user or an error.
// Token generation is handled by the presentation layer.
func (uc *LoginUseCase) Execute(ctx context.Context, input dtos.LoginInputDTO) (*presentation_dtos.AuthResponseDTO, error) {
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

func (uc *LoginUseCase) buildLoginResponse(user *domain.User, token string) presentation_dtos.AuthResponseDTO {
	userResponse := presentation_dtos.UserResponseDTO{
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
	return presentation_dtos.AuthResponseDTO{
		User:  userResponse,
		Token: token,
	}
}
