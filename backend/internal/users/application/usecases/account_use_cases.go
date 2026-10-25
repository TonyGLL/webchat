package usecases

import (
	"context"
	"io"

	auth_domain "backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/users/domain"
)

// ExportAccountUseCase builds the "download my data" file.
type ExportAccountUseCase struct {
	usersRepo domain.UsersRepository
}

func NewExportAccountUseCase(usersRepo domain.UsersRepository) *ExportAccountUseCase {
	return &ExportAccountUseCase{usersRepo: usersRepo}
}

func (uc *ExportAccountUseCase) Execute(ctx context.Context, userID int, w io.Writer) error {
	return uc.usersRepo.ExportAccount(ctx, userID, w)
}

// Realtime lets account deletion cut the user's open websocket connections.
type Realtime interface {
	DisconnectUser(userID int)
}

// DeleteAccountUseCase permanently deletes an account after the user confirms their password.
type DeleteAccountUseCase struct {
	usersRepo domain.UsersRepository
	passwords auth_domain.PasswordService
	realtime  Realtime
}

func NewDeleteAccountUseCase(usersRepo domain.UsersRepository, passwords auth_domain.PasswordService, realtime Realtime) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{usersRepo: usersRepo, passwords: passwords, realtime: realtime}
}

func (uc *DeleteAccountUseCase) Execute(ctx context.Context, userID int, password string) error {
	hash, err := uc.usersRepo.PasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	if !uc.passwords.CheckPasswordHash(password, hash) {
		return shared_domain.ErrInvalidCredentials
	}
	if err := uc.usersRepo.DeleteAccount(ctx, userID); err != nil {
		return err
	}
	if uc.realtime != nil {
		uc.realtime.DisconnectUser(userID)
	}
	return nil
}
