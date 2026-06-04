package handler

import (
	"getkanbam.app/api/internal/convert"
	"github.com/jackc/pgx/v5/pgtype"
)

// parseQueryUUID parses an optional query string UUID parameter.
// Returns a zero-value (Valid: false) UUID if empty or invalid.
func parseQueryUUID(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{}
	}
	id, err := convert.ParseUUID(s)
	if err != nil {
		return pgtype.UUID{}
	}
	return id
}

// parseQueryTime parses an optional RFC3339 query string timestamp.
// Returns a zero-value (Valid: false) Timestamptz if empty or invalid.
func parseQueryTime(s string) pgtype.Timestamptz {
	return convert.ParseTimestamptz(s)
}
