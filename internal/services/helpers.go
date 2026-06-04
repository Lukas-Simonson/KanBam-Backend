package services

import (
	"strings"

	"getkanbam.app/api/internal/convert"
	"getkanbam.app/api/internal/db"
	"getkanbam.app/api/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// ActivityFilters holds optional filter parameters for activity queries.
type ActivityFilters struct {
	BoardID pgtype.UUID
	CardID  pgtype.UUID
	UserID  pgtype.UUID
	After   pgtype.Timestamptz
	Before  pgtype.Timestamptz
}

// isUniqueViolation checks if a Postgres error is a unique constraint violation.
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "23505")
}

// tagToModel converts a db.Tag to model.Tag.
func tagToModel(t db.Tag) model.Tag {
	return model.Tag{
		ID:          convert.UUIDToString(t.ID),
		Name:        t.Name,
		Color:       convert.ColorToString(t.Color),
		WorkspaceID: convert.UUIDToString(t.WorkspaceID),
		CreatedAt:   convert.TimeToString(t.CreatedAt),
		UpdatedAt:   convert.NullableTimeToString(t.UpdatedAt),
	}
}

// activityToModel converts a db.Activity to model.Activity.
func activityToModel(a db.Activity) model.Activity {
	return model.Activity{
		ID:          convert.UUIDToString(a.ID),
		WorkspaceID: convert.UUIDToString(a.WorkspaceID),
		Description: a.Description,
		CreatedAt:   convert.TimeToString(a.CreatedAt),
		UserID:      convert.NullableUUIDToString(a.UserID),
		BoardID:     convert.NullableUUIDToString(a.BoardID),
		ColumnID:    convert.NullableUUIDToString(a.ColumnID),
		CardID:      convert.NullableUUIDToString(a.CardID),
		CommentID:   convert.NullableUUIDToString(a.CommentID),
		TagID:       convert.NullableUUIDToString(a.TagID),
	}
}

func activitiesToModel(rows []db.Activity) []model.Activity {
	result := make([]model.Activity, len(rows))
	for i, a := range rows {
		result[i] = activityToModel(a)
	}
	return result
}
