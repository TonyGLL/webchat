package domain

import (
	"context"
)

type AuthRepository interface {
	GetUserByID(ctx context.Context, id int) (*User, error)
	GetUserByEmailOrUsername(ctx context.Context, emailOrUsername string) (*User, error)
	GetUserVerifyEmail(ctx context.Context, email string, userID int) (*User, error)
	ValidateUserPassword(ctx context.Context, id int) (string, error)
	Register(ctx context.Context, user *User) (*User, error)
	SetLastAccess(ctx context.Context, id int) error
	CreatePassword(ctx context.Context, userID int, hashedPassword string) error
}
