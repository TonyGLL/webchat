package domain

import (
	"context"
)

type AuthRepository interface {
	GetUserByID(ctx context.Context, id int) (*User, error)
	GetUserByEmailOrUsername(ctx context.Context, emailOrUsername string) (*User, error)
	GetUserVerifyEmail(ctx context.Context, email string, userID int) (*User, error)
	ValidateUserPassword(ctx context.Context, id int) (string, error)
	// RegisterWithPassword creates the user and their password atomically.
	RegisterWithPassword(ctx context.Context, user *User, hashedPassword string) (*User, error)
	SetLastAccess(ctx context.Context, id int) error
	VerifyEmail(ctx context.Context, id int) error
	// UpdatePassword replaces the stored hash of an existing user.
	UpdatePassword(ctx context.Context, id int, hashedPassword string) error
}
