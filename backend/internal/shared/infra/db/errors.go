package db

import (
	"errors"

	"github.com/lib/pq"
)

// PostgreSQL error codes (https://www.postgresql.org/docs/current/errcodes-appendix.html).
const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
)

func hasCode(err error, code string) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == code
}

// IsUniqueViolation reports whether err is a violated UNIQUE / PRIMARY KEY constraint.
func IsUniqueViolation(err error) bool { return hasCode(err, codeUniqueViolation) }

// IsForeignKeyViolation reports whether err is a violated FOREIGN KEY constraint.
func IsForeignKeyViolation(err error) bool { return hasCode(err, codeForeignKeyViolation) }
