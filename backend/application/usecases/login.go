package usecases

import (
	"backend/application"
	"backend/application/dtos"
	"backend/domain"
	"context"
	"errors"
)

type LoginUseCase struct {
	store           application.Store
	passwordService domain.PasswordService
}

func NewLoginUseCase(store application.Store, ps domain.PasswordService) *LoginUseCase {
	return &LoginUseCase{
		store:           store,
		passwordService: ps,
	}
}

// Execute handles user login. It returns the authenticated user or an error.
// Token generation is handled by the presentation layer.
func (uc *LoginUseCase) Execute(ctx context.Context, input dtos.LoginInputDTO) (*domain.User, error) {
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

	return user, nil
}