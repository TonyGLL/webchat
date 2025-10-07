package repositories

import (
	"backend/application/dtos"
	"backend/domain"
	"context"
)

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, arg *dtos.GetUserByEmailDTO) (*domain.User, error)
	ValidateUserPasword(ctx context.Context, id int) (string, error)
	Register(user *domain.User, password string) (*domain.User, error)
}
