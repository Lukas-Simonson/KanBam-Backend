package services

import (
	"context"
	"errors"
	"fmt"

	"getkanbam.app/api/internal/apierr"
	"getkanbam.app/api/internal/convert"
	"getkanbam.app/api/internal/db"
	"getkanbam.app/api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type WorkspaceService struct {
	db *db.Queries
}

func NewWorkspaceService(q *db.Queries) WorkspaceService {
	return WorkspaceService{db: q}
}

// --- Workspaces ---

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID string) ([]model.Workspace, error) {
	id, err := convert.ParseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}
	rows, err := s.db.GetWorkspacesByUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}
	result := make([]model.Workspace, len(rows))
	for i, w := range rows {
		result[i] = workspaceToModel(w)
	}
	return result, nil
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, callerID, workspaceID string) (model.Workspace, error) {
	id, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return model.Workspace{}, fmt.Errorf("parsing workspace id: %w", err)
	}
	w, err := s.db.GetWorkspaceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Workspace{}, apierr.WorkspaceNotFound()
		}
		return model.Workspace{}, fmt.Errorf("getting workspace: %w", err)
	}
	if err := requireRole(ctx, s.db, id, callerID, db.RoleViewer); err != nil {
		return model.Workspace{}, err
	}
	return workspaceToModel(w), nil
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, userID string, req model.WorkspaceCreation) (model.Workspace, error) {
	ownerID, err := convert.ParseUUID(userID)
	if err != nil {
		return model.Workspace{}, fmt.Errorf("parsing user id: %w", err)
	}
	w, err := s.db.CreateWorkspace(ctx, db.CreateWorkspaceParams{
		ID:          convert.NewUUID(),
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     ownerID,
	})
	if err != nil {
		return model.Workspace{}, fmt.Errorf("creating workspace: %w", err)
	}
	return workspaceToModel(w), nil
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, callerID, workspaceID string, req model.WorkspaceUpdate) (model.Workspace, error) {
	id, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return model.Workspace{}, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, id, callerID, db.RoleAdmin); err != nil {
		return model.Workspace{}, err
	}
	current, err := s.db.GetWorkspaceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Workspace{}, apierr.WorkspaceNotFound()
		}
		return model.Workspace{}, fmt.Errorf("getting workspace: %w", err)
	}

	title := current.Title
	if req.Title != nil {
		title = *req.Title
	}
	description := current.Description
	if req.Description != nil {
		description = req.Description
	}

	w, err := s.db.UpdateWorkspace(ctx, db.UpdateWorkspaceParams{
		ID:          id,
		Title:       title,
		Description: description,
	})
	if err != nil {
		return model.Workspace{}, fmt.Errorf("updating workspace: %w", err)
	}
	return workspaceToModel(w), nil
}

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, callerID, workspaceID string) error {
	id, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, id, callerID, db.RoleOwner); err != nil {
		return err
	}
	if err := s.db.DeleteWorkspace(ctx, id); err != nil {
		return fmt.Errorf("deleting workspace: %w", err)
	}
	return nil
}

// --- Members ---

func (s *WorkspaceService) ListMembers(ctx context.Context, callerID, workspaceID string) ([]model.UserRole, error) {
	id, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, id, callerID, db.RoleViewer); err != nil {
		return nil, err
	}
	rows, err := s.db.GetWorkspaceMembers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	result := make([]model.UserRole, len(rows))
	for i, r := range rows {
		result[i] = model.UserRole{
			User: model.User{
				ID:        convert.UUIDToString(r.ID),
				Name:      r.Name,
				Email:     r.Email,
				CreatedAt: convert.TimeToString(r.CreatedAt),
			},
			Role: string(r.Role),
		}
	}
	return result, nil
}

func (s *WorkspaceService) GetMember(ctx context.Context, callerID, workspaceID, userID string) (model.UserRole, error) {
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return model.UserRole{}, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, wsID, callerID, db.RoleViewer); err != nil {
		return model.UserRole{}, err
	}
	uID, err := convert.ParseUUID(userID)
	if err != nil {
		return model.UserRole{}, fmt.Errorf("parsing user id: %w", err)
	}
	r, err := s.db.GetWorkspaceMember(ctx, db.GetWorkspaceMemberParams{
		WorkspaceID: wsID,
		UserID:      uID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserRole{}, apierr.NotAuthorized()
		}
		return model.UserRole{}, fmt.Errorf("getting member: %w", err)
	}
	return model.UserRole{
		User: model.User{
			ID:        convert.UUIDToString(r.ID),
			Name:      r.Name,
			Email:     r.Email,
			CreatedAt: convert.TimeToString(r.CreatedAt),
		},
		Role: string(r.Role),
	}, nil
}

func (s *WorkspaceService) AddMember(ctx context.Context, callerID, workspaceID string, req model.Membership) error {
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, wsID, callerID, db.RoleAdmin); err != nil {
		return err
	}
	uID, err := convert.ParseUUID(req.UserID)
	if err != nil {
		return fmt.Errorf("parsing user id: %w", err)
	}

	_, err = s.db.GetWorkspaceMembership(ctx, db.GetWorkspaceMembershipParams{
		WorkspaceID: wsID,
		UserID:      uID,
	})
	if err == nil {
		return apierr.AlreadyInWorkspace()
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("checking membership: %w", err)
	}

	if err := s.db.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		UserID:      uID,
		WorkspaceID: wsID,
		Role:        db.Role(req.Role),
	}); err != nil {
		return fmt.Errorf("adding member: %w", err)
	}
	return nil
}

func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, callerID, workspaceID, userID string, req model.MembershipUpdate) error {
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, wsID, callerID, db.RoleOwner); err != nil {
		return err
	}
	uID, err := convert.ParseUUID(userID)
	if err != nil {
		return fmt.Errorf("parsing user id: %w", err)
	}
	if err := s.db.UpdateWorkspaceMemberRole(ctx, db.UpdateWorkspaceMemberRoleParams{
		WorkspaceID: wsID,
		UserID:      uID,
		Role:        db.Role(req.Role),
	}); err != nil {
		return fmt.Errorf("updating member role: %w", err)
	}
	return nil
}

func (s *WorkspaceService) RemoveMember(ctx context.Context, callerID, workspaceID, userID string) error {
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return fmt.Errorf("parsing workspace id: %w", err)
	}

	// Any member can remove themselves; removing someone else requires admin.
	if callerID == userID {
		if err := requireRole(ctx, s.db, wsID, callerID, db.RoleViewer); err != nil {
			return err
		}
	} else {
		if err := requireRole(ctx, s.db, wsID, callerID, db.RoleAdmin); err != nil {
			return err
		}
	}

	uID, err := convert.ParseUUID(userID)
	if err != nil {
		return fmt.Errorf("parsing user id: %w", err)
	}
	if err := s.db.RemoveWorkspaceMember(ctx, db.RemoveWorkspaceMemberParams{
		WorkspaceID: wsID,
		UserID:      uID,
	}); err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	return nil
}

// --- Tags (workspace-scoped listing + creation) ---

func (s *WorkspaceService) ListTags(ctx context.Context, callerID, workspaceID string) ([]model.Tag, error) {
	id, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, id, callerID, db.RoleViewer); err != nil {
		return nil, err
	}
	rows, err := s.db.GetTagsByWorkspace(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	result := make([]model.Tag, len(rows))
	for i, t := range rows {
		result[i] = tagToModel(t)
	}
	return result, nil
}

func (s *WorkspaceService) CreateTag(ctx context.Context, callerID, workspaceID string, req model.TagCreation) (model.Tag, error) {
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return model.Tag{}, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, wsID, callerID, db.RoleContributer); err != nil {
		return model.Tag{}, err
	}
	t, err := s.db.CreateTag(ctx, db.CreateTagParams{
		ID:          convert.NewUUID(),
		Name:        req.Name,
		Color:       req.Color,
		WorkspaceID: wsID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return model.Tag{}, apierr.TagNameTaken()
		}
		return model.Tag{}, fmt.Errorf("creating tag: %w", err)
	}
	return tagToModel(t), nil
}

// --- Activity ---

func (s *WorkspaceService) GetActivity(ctx context.Context, callerID, workspaceID string, filters ActivityFilters) ([]model.Activity, error) {
	id, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, id, callerID, db.RoleViewer); err != nil {
		return nil, err
	}
	rows, err := s.db.GetWorkspaceActivity(ctx, db.GetWorkspaceActivityParams{
		WorkspaceID: id,
		Column2:     filters.BoardID,
		Column3:     filters.CardID,
		Column4:     filters.UserID,
		Column5:     filters.After,
		Column6:     filters.Before,
	})
	if err != nil {
		return nil, fmt.Errorf("getting workspace activity: %w", err)
	}
	return activitiesToModel(rows), nil
}

// --- Helpers ---

func workspaceToModel(w db.Workspace) model.Workspace {
	return model.Workspace{
		ID:          convert.UUIDToString(w.ID),
		Title:       w.Title,
		Description: w.Description,
		OwnerID:     convert.UUIDToString(w.OwnerID),
		CreatedAt:   convert.TimeToString(w.CreatedAt),
		UpdatedAt:   convert.NullableTimeToString(w.UpdatedAt),
	}
}

func nullableUUID(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{}
	}
	id, err := convert.ParseUUID(*s)
	if err != nil {
		return pgtype.UUID{}
	}
	return id
}
