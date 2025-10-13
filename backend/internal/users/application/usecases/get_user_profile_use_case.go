package usecases

import (
	shared_domain "backend/internal/shared/domain"
	"backend/internal/users/application/dtos"
	"backend/internal/users/domain"
	"context"
)

type GetUserProfileUseCase struct {
	usersRepo domain.UsersRepository
}

func NewGetUserProfileUseCase(usersRepo domain.UsersRepository) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{
		usersRepo: usersRepo,
	}
}

func (uc *GetUserProfileUseCase) Execute(ctx context.Context, userID int) (*dtos.UserProfileDTO, error) {
	user, err := uc.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, shared_domain.ErrNotFound
	}

	return &dtos.UserProfileDTO{
		ID:        user.ID,
		Name:      user.Name,
		LastName:  user.LastName,
		Username:  user.Username,
		AvatarUrl: user.AvatarUrl,
		CreatedAt: user.CreatedAt,
	}, nil
}
