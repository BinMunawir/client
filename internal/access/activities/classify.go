package access_activities

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// isUniqueViolation classifies a Postgres unique-constraint violation at the activity
// boundary (standard §6.2), so the stores below stay dumb and callers stay domain-only.
func isUniqueViolation(err error) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == pgerrcode.UniqueViolation
}
