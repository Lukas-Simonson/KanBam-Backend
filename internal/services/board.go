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
)

type BoardService struct {
	db *db.Queries
}

func NewBoardService(q *db.Queries) BoardService {
	return BoardService{db: q}
}

func (s *BoardService) ListBoards(ctx context.Context, callerID, workspaceID string) ([]model.Board, error) {
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, wsID, callerID, db.RoleViewer); err != nil {
		return nil, err
	}
	rows, err := s.db.GetBoardsByWorkspace(ctx, wsID)
	if err != nil {
		return nil, fmt.Errorf("listing boards: %w", err)
	}
	result := make([]model.Board, len(rows))
	for i, b := range rows {
		result[i] = boardToModel(b)
	}
	return result, nil
}

func (s *BoardService) GetBoard(ctx context.Context, callerID, boardID string) (model.Board, error) {
	id, err := convert.ParseUUID(boardID)
	if err != nil {
		return model.Board{}, fmt.Errorf("parsing board id: %w", err)
	}
	b, err := s.db.GetBoardByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Board{}, apierr.BoardNotFound()
		}
		return model.Board{}, fmt.Errorf("getting board: %w", err)
	}
	if err := requireRole(ctx, s.db, b.WorkspaceID, callerID, db.RoleViewer); err != nil {
		return model.Board{}, err
	}
	return boardToModel(b), nil
}

func (s *BoardService) CreateBoard(ctx context.Context, callerID, workspaceID string, req model.BoardCreation) (model.Board, error) {
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return model.Board{}, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, wsID, callerID, db.RoleContributer); err != nil {
		return model.Board{}, err
	}
	b, err := s.db.CreateBoard(ctx, db.CreateBoardParams{
		ID:          convert.NewUUID(),
		Title:       req.Title,
		Description: req.Description,
		Prefix:      req.Prefix,
		WorkspaceID: wsID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return model.Board{}, apierr.BoardPrefixTaken()
		}
		return model.Board{}, fmt.Errorf("creating board: %w", err)
	}
	return boardToModel(b), nil
}

func (s *BoardService) UpdateBoard(ctx context.Context, callerID, boardID string, req model.BoardUpdate) (model.Board, error) {
	id, err := convert.ParseUUID(boardID)
	if err != nil {
		return model.Board{}, fmt.Errorf("parsing board id: %w", err)
	}
	current, err := s.db.GetBoardByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Board{}, apierr.BoardNotFound()
		}
		return model.Board{}, fmt.Errorf("getting board: %w", err)
	}
	if err := requireRole(ctx, s.db, current.WorkspaceID, callerID, db.RoleAdmin); err != nil {
		return model.Board{}, err
	}

	title := current.Title
	if req.Title != nil {
		title = *req.Title
	}
	description := current.Description
	if req.Description != nil {
		description = req.Description
	}
	isArchived := current.IsArchived
	if req.IsArchived != nil {
		isArchived = *req.IsArchived
	}

	b, err := s.db.UpdateBoard(ctx, db.UpdateBoardParams{
		ID:          id,
		Title:       title,
		Description: description,
		IsArchived:  isArchived,
	})
	if err != nil {
		return model.Board{}, fmt.Errorf("updating board: %w", err)
	}
	return boardToModel(b), nil
}

func (s *BoardService) DeleteBoard(ctx context.Context, callerID, boardID string) error {
	id, err := convert.ParseUUID(boardID)
	if err != nil {
		return fmt.Errorf("parsing board id: %w", err)
	}
	b, err := s.db.GetBoardByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierr.BoardNotFound()
		}
		return fmt.Errorf("getting board: %w", err)
	}
	if err := requireRole(ctx, s.db, b.WorkspaceID, callerID, db.RoleAdmin); err != nil {
		return err
	}
	if err := s.db.DeleteBoard(ctx, id); err != nil {
		return fmt.Errorf("deleting board: %w", err)
	}
	return nil
}

func (s *BoardService) GetActivity(ctx context.Context, callerID, boardID string, workspaceID string, filters ActivityFilters) ([]model.Activity, error) {
	bID, err := convert.ParseUUID(boardID)
	if err != nil {
		return nil, fmt.Errorf("parsing board id: %w", err)
	}
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("parsing workspace id: %w", err)
	}
	if err := requireRole(ctx, s.db, wsID, callerID, db.RoleViewer); err != nil {
		return nil, err
	}
	rows, err := s.db.GetBoardActivity(ctx, db.GetBoardActivityParams{
		WorkspaceID: wsID,
		BoardID:     bID,
		Column3:     filters.CardID,
		Column4:     filters.UserID,
		Column5:     filters.After,
		Column6:     filters.Before,
	})
	if err != nil {
		return nil, fmt.Errorf("getting board activity: %w", err)
	}
	return activitiesToModel(rows), nil
}

func boardToModel(b db.Board) model.Board {
	return model.Board{
		ID:          convert.UUIDToString(b.ID),
		Title:       b.Title,
		Description: b.Description,
		Prefix:      b.Prefix,
		Sequence:    b.Sequence,
		IsArchived:  b.IsArchived,
		WorkspaceID: convert.UUIDToString(b.WorkspaceID),
		CreatedAt:   convert.TimeToString(b.CreatedAt),
		UpdatedAt:   convert.NullableTimeToString(b.UpdatedAt),
	}
}
