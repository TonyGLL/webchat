package usecases

import (
	"backend/application/dtos"
	"backend/application/repositories"
	"backend/domain"
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	AuthRepository repositories.AuthRepository
}

func NewLoginUseCase(repo repositories.AuthRepository) *LoginUseCase {
	return &LoginUseCase{AuthRepository: repo}
}

// Execute handles user login. It returns the authenticated user or an error.
// Token generation is handled by the presentation layer.
func (uc *LoginUseCase) Execute(ctx context.Context, input dtos.LoginInputDTO) (*domain.User, error) {
	user, err := uc.AuthRepository.GetUserByEmailOrUsername(ctx, &dtos.GetUserByEmailOrUsernameDTO{User: input.User})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvalidCredentials // User not found
		}
		return nil, err
	}

	hashedPassword, err := uc.AuthRepository.ValidateUserPasword(ctx, user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvalidCredentials // Password not found for user
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials // Passwords do not match
	}

	if err := uc.AuthRepository.SetLastAccess(ctx, user.ID); err != nil {
		// This error might not be critical for the login flow itself.
		// For now, we return it, but it could also just be logged.
		return nil, err
	}

	return user, nil
}
