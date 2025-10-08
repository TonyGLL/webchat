package repositories

import (
	"backend/domain"
	"context"
)

type AuthRepository interface {
	GetUserByEmailOrUsername(ctx context.Context, emailOrUsername string) (*domain.User, error)
	ValidateUserPassword(ctx context.Context, id int) (string, error)
	Register(ctx context.Context, user *domain.User) (*domain.User, error)
	SetLastAccess(ctx context.Context, id int) error
	CreatePassword(ctx context.Context, userID int, hashedPassword string) error
}
