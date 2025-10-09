package usecases

import (
	auth_domain "backend/internal/auth/domain"
	"context"
)

type DeactivateUserUseCase struct {
	authRepo auth_domain.AuthRepository
}

func NewDeactivateUserUseCase(
	authRepo auth_domain.AuthRepository) *DeactivateUserUseCase {
	return &DeactivateUserUseCase{
		authRepo: authRepo,
	}
}

// Execute handles user login. It returns the authenticated user or an error.
func (uc *DeactivateUserUseCase) Execute(ctx context.Context, input any) error {
	return nil
}
