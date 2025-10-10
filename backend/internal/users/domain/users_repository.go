package domain

import (
	auth_domain "backend/internal/auth/domain"
	"context"
)

type UsersRepository interface {
	GetUserByID(ctx context.Context, id int) (*auth_domain.User, error)
	DeactivateUser(ctx context.Context, id int) error
}
