package domain

import "context"

// TransactionManager defines the interface for managing database transactions.
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
