package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"getkanbam.app/api/internal/apierr"
	"getkanbam.app/api/internal/convert"
	"getkanbam.app/api/internal/db"
	"getkanbam.app/api/internal/model"
	"github.com/jackc/pgx/v5"
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

// roleAtLeast returns true when role meets or exceeds minRole in the hierarchy:
// viewer < contributer < admin < owner.
func roleAtLeast(role, minRole db.Role) bool {
	order := map[db.Role]int{
		db.RoleViewer:      0,
		db.RoleContributer: 1,
		db.RoleAdmin:       2,
		db.RoleOwner:       3,
	}
	return order[role] >= order[minRole]
}

// requireRole verifies callerID is an active workspace member with at least minRole.
// Returns NotAuthorized if the user is not a member, has a pending invite, or lacks the required role.
func requireRole(ctx context.Context, q *db.Queries, workspaceID pgtype.UUID, callerID string, minRole db.Role) error {
	uID, err := convert.ParseUUID(callerID)
	if err != nil {
		return apierr.NotAuthorized()
	}
	m, err := q.GetWorkspaceMembership(ctx, db.GetWorkspaceMembershipParams{
		WorkspaceID: workspaceID,
		UserID:      uID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierr.NotAuthorized()
		}
		return fmt.Errorf("checking membership: %w", err)
	}
	if m.Pending || !roleAtLeast(m.Role, minRole) {
		return apierr.NotAuthorized()
	}
	return nil
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
