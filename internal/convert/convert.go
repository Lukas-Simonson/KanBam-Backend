package convert

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// NewUUID creates a new random pgtype.UUID.
func NewUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.New(), Valid: true}
}

// ParseUUID parses a UUID string into a pgtype.UUID.
func ParseUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

// UUIDToString converts a pgtype.UUID to its string representation.
func UUIDToString(id pgtype.UUID) string {
	return uuid.UUID(id.Bytes).String()
}

// NullableUUIDToString converts a nullable pgtype.UUID to *string; returns nil if not valid.
func NullableUUIDToString(id pgtype.UUID) *string {
	if !id.Valid {
		return nil
	}
	s := uuid.UUID(id.Bytes).String()
	return &s
}

// TimeToString formats a pgtype.Timestamptz as RFC3339.
func TimeToString(t pgtype.Timestamptz) string {
	return t.Time.Format(time.RFC3339)
}

// NullableTimeToString formats a nullable pgtype.Timestamptz as *string; returns nil if not valid.
func NullableTimeToString(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

// ParseTimestamptz parses an RFC3339 string into a pgtype.Timestamptz.
// Returns a zero-value (Valid: false) if the input is empty.
func ParseTimestamptz(s string) pgtype.Timestamptz {
	if s == "" {
		return pgtype.Timestamptz{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// ColorToString extracts the hex_color domain value (stored as interface{}) to string.
func ColorToString(v interface{}) string {
	s, _ := v.(string)
	return s
}
