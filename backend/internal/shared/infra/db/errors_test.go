package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
)

func TestPostgresErrorClassification(t *testing.T) {
	unique := fmt.Errorf("wrapped: %w", &pq.Error{Code: "23505"})
	fk := &pq.Error{Code: "23503"}

	if !IsUniqueViolation(unique) || IsForeignKeyViolation(unique) {
		t.Fatal("23505 must be a unique violation only")
	}
	if !IsForeignKeyViolation(fk) || IsUniqueViolation(fk) {
		t.Fatal("23503 must be a foreign key violation only")
	}
	if IsUniqueViolation(errors.New("boom")) || IsUniqueViolation(nil) {
		t.Fatal("other errors are not violations")
	}
}
