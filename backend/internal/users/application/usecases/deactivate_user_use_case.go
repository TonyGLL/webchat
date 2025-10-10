package usecases

import (
	shared_domain "backend/internal/shared/domain"
	"backend/internal/users/domain"
	"context"
)

type DeactivateUserUseCase struct {
	usersRepo domain.UsersRepository
}

func NewDeactivateUserUseCase(
	usersRepo domain.UsersRepository) *DeactivateUserUseCase {
	return &DeactivateUserUseCase{
		usersRepo: usersRepo,
	}
}

// Execute handles user login. It returns the authenticated user or an error.
func (uc *DeactivateUserUseCase) Execute(ctx context.Context, userID int) error {
	user, err := uc.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return shared_domain.ErrNotFound
	}

	err = uc.usersRepo.DeactivateUser(ctx, user.ID)
	if err != nil {
		return err
	}

	return nil
}
