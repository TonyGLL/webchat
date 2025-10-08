package domain

import (
	"context"
)

type AuthRepository interface {
	GetUserByEmailOrUsername(ctx context.Context, emailOrUsername string) (*User, error)
	ValidateUserPassword(ctx context.Context, id int) (string, error)
	Register(ctx context.Context, user *User) (*User, error)
	SetLastAccess(ctx context.Context, id int) error
	CreatePassword(ctx context.Context, userID int, hashedPassword string) error
}