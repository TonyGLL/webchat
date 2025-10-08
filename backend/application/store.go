package application

import (
	"backend/application/repositories"
	"context"
)

// Store defines the interface for a unit of work, providing access to repositories
// and a mechanism to run operations within a database transaction.
type Store interface {
	// AuthRepository returns an implementation of AuthRepository.
	AuthRepository() repositories.AuthRepository

	// MessageRepository returns an implementation of MessageRepository.
	MessageRepository() repositories.MessageRepository

	// ExecTx executes a function within a database transaction.
	// It commits the transaction if the function returns no error,

	// and rolls it back otherwise.
	ExecTx(ctx context.Context, fn func(Store) error) error
}