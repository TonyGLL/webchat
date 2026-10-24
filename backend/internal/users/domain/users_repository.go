package domain

import (
	auth_domain "backend/internal/auth/domain"
	"context"
)

type UsersRepository interface {
	GetUserByID(ctx context.Context, id int) (*auth_domain.User, error)
	DeactivateUser(ctx context.Context, id int) error
	// PasswordHash returns the stored password hash, or ErrNotFound.
	PasswordHash(ctx context.Context, id int) (string, error)
	ExportAccount(ctx context.Context, id int) (*AccountExport, error)
	// DeleteAccount permanently removes the user. Rooms they own pass to the longest-standing
	// admin (else member); rooms with nobody else are deleted; their messages stay, anonymised.
	DeleteAccount(ctx context.Context, id int) error
	// ListDirectory returns active users other than selfID whose name or username contains q.
	ListDirectory(ctx context.Context, selfID int, q string, limit int) ([]*DirectoryEntry, error)
}
