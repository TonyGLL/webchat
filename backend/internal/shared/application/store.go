package application

import (
	"context"
)

// Store defines the interface for a unit of work, providing a mechanism to run operations within a database transaction.
type Store interface {
	// ExecTx executes a function within a database transaction.
	// It commits the transaction if the function returns no error,
	// and rolls it back otherwise.
	ExecTx(ctx context.Context, fn func(Store) error) error
}
