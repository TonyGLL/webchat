package db

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/shared/application"
)

// SQLStore provides an implementation of the application.Store interface for PostgreSQL.
// It manages database transactions.
type SQLStore struct {
	db DBTX
}

// NewSQLStore creates a new SQLStore.
func NewSQLStore(db DBTX) application.Store {
	return &SQLStore{
		db: db,
	}
}

// ExecTx executes a function within a database transaction.
func (s *SQLStore) ExecTx(ctx context.Context, fn func(store application.Store) error) error {
	tx, err := s.db.(*sql.DB).BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Create a new store instance with the transaction.
	txStore := &SQLStore{
		db: tx,
	}

	err = fn(txStore)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
