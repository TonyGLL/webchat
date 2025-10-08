package persistence

import (
	"backend/application"
	"backend/application/repositories"
	"context"
	"database/sql"
	"fmt"
)

// SQLStore provides an implementation of the application.Store interface for PostgreSQL.
// It manages database transactions and provides access to repositories.
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore creates a new SQLStore.
func NewSQLStore(db *sql.DB) application.Store {
	return &SQLStore{
		db: db,
	}
}

// AuthRepository returns an implementation of AuthRepository.
func (s *SQLStore) AuthRepository() repositories.AuthRepository {
	return NewPgAuthRepository(s.db)
}

// MessageRepository returns an implementation of MessageRepository.
func (s *SQLStore) MessageRepository() repositories.MessageRepository {
	return NewPgMessageRepository(s.db)
}

// ExecTx executes a function within a database transaction.
func (s *SQLStore) ExecTx(ctx context.Context, fn func(store application.Store) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Create a new store instance with the transaction.
	txStore := &txStore{
		tx: tx,
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

// txStore is a store that operates within a transaction.
type txStore struct {
	tx *sql.Tx
}

func (s *txStore) AuthRepository() repositories.AuthRepository {
	return NewPgAuthRepository(s.tx)
}

func (s *txStore) MessageRepository() repositories.MessageRepository {
	return NewPgMessageRepository(s.tx)
}

func (s *txStore) ExecTx(ctx context.Context, fn func(store application.Store) error) error {
	// Nested transactions are not supported in this implementation.
	return fn(s)
}