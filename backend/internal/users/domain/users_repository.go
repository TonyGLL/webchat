package domain

import (
	"context"
)

type UsersRepository interface {
	DeactivateUser(ctx context.Context, id int) error
}
