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

type ColumnService struct {
	db *db.Queries
}

func NewColumnService(q *db.Queries) ColumnService {
	return ColumnService{db: q}
}

func (s *ColumnService) GetColumn(ctx context.Context, callerID, columnID string) (model.Column, error) {
	id, err := convert.ParseUUID(columnID)
	if err != nil {
		return model.Column{}, fmt.Errorf("parsing column id: %w", err)
	}
	col, err := s.db.GetColumnByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Column{}, apierr.ColumnNotFound()
		}
		return model.Column{}, fmt.Errorf("getting column: %w", err)
	}
	board, err := s.db.GetBoardByID(ctx, col.BoardID)
	if err != nil {
		return model.Column{}, fmt.Errorf("getting board for column: %w", err)
	}
	if err := requireRole(ctx, s.db, board.WorkspaceID, callerID, db.RoleViewer); err != nil {
		return model.Column{}, err
	}
	return columnToModel(col), nil
}

// ListColumnsWithCards returns all columns in a board, each populated with its cards.
func (s *ColumnService) ListColumnsWithCards(ctx context.Context, callerID, boardID string) ([]model.ColumnWithCards, error) {
	bID, err := convert.ParseUUID(boardID)
	if err != nil {
		return nil, fmt.Errorf("parsing board id: %w", err)
	}
	board, err := s.db.GetBoardByID(ctx, bID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierr.BoardNotFound()
		}
		return nil, fmt.Errorf("getting board: %w", err)
	}
	if err := requireRole(ctx, s.db, board.WorkspaceID, callerID, db.RoleViewer); err != nil {
		return nil, err
	}
	cols, err := s.db.GetColumnsByBoard(ctx, bID)
	if err != nil {
		return nil, fmt.Errorf("listing columns: %w", err)
	}

	result := make([]model.ColumnWithCards, len(cols))
	for i, col := range cols {
		cards, err := s.db.GetCardsByBoard(ctx, bID)
		if err != nil {
			return nil, fmt.Errorf("listing cards for column: %w", err)
		}
		columnCards := make([]model.Card, 0)
		for _, c := range cards {
			if c.ColumnID.Valid && convert.UUIDToString(c.ColumnID) == convert.UUIDToString(col.ID) {
				columnCards = append(columnCards, cardToModel(c))
			}
		}
		result[i] = model.ColumnWithCards{
			Column: columnToModel(col),
			Cards:  columnCards,
		}
	}
	return result, nil
}

func (s *ColumnService) CreateColumn(ctx context.Context, callerID, boardID string, req model.ColumnCreation) (model.Column, error) {
	bID, err := convert.ParseUUID(boardID)
	if err != nil {
		return model.Column{}, fmt.Errorf("parsing board id: %w", err)
	}
	board, err := s.db.GetBoardByID(ctx, bID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Column{}, apierr.BoardNotFound()
		}
		return model.Column{}, fmt.Errorf("getting board: %w", err)
	}
	if err := requireRole(ctx, s.db, board.WorkspaceID, callerID, db.RoleContributer); err != nil {
		return model.Column{}, err
	}
	col, err := s.db.CreateColumn(ctx, db.CreateColumnParams{
		ID:         convert.NewUUID(),
		Title:      req.Title,
		WipLimit:   req.WipLimit,
		Color:      req.Color,
		IsTerminal: req.IsTerminal,
		Position:   req.Position,
		BoardID:    bID,
	})
	if err != nil {
		return model.Column{}, fmt.Errorf("creating column: %w", err)
	}
	return columnToModel(col), nil
}

func (s *ColumnService) UpdateColumn(ctx context.Context, callerID, columnID string, req model.ColumnUpdate) (model.Column, error) {
	id, err := convert.ParseUUID(columnID)
	if err != nil {
		return model.Column{}, fmt.Errorf("parsing column id: %w", err)
	}
	current, err := s.db.GetColumnByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Column{}, apierr.ColumnNotFound()
		}
		return model.Column{}, fmt.Errorf("getting column: %w", err)
	}
	board, err := s.db.GetBoardByID(ctx, current.BoardID)
	if err != nil {
		return model.Column{}, fmt.Errorf("getting board for column: %w", err)
	}
	if err := requireRole(ctx, s.db, board.WorkspaceID, callerID, db.RoleAdmin); err != nil {
		return model.Column{}, err
	}

	title := current.Title
	if req.Title != nil {
		title = *req.Title
	}
	wipLimit := current.WipLimit
	if req.WipLimit != nil {
		wipLimit = req.WipLimit
	}
	isTerminal := current.IsTerminal
	if req.IsTerminal != nil {
		isTerminal = *req.IsTerminal
	}
	position := current.Position
	if req.Position != nil {
		position = *req.Position
	}
	var color interface{} = current.Color
	if req.Color != nil {
		color = *req.Color
	}

	col, err := s.db.UpdateColumn(ctx, db.UpdateColumnParams{
		ID:         id,
		Title:      title,
		WipLimit:   wipLimit,
		Color:      color,
		IsTerminal: isTerminal,
		Position:   position,
	})
	if err != nil {
		return model.Column{}, fmt.Errorf("updating column: %w", err)
	}
	return columnToModel(col), nil
}

func (s *ColumnService) DeleteColumn(ctx context.Context, callerID, columnID string) error {
	id, err := convert.ParseUUID(columnID)
	if err != nil {
		return fmt.Errorf("parsing column id: %w", err)
	}
	col, err := s.db.GetColumnByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierr.ColumnNotFound()
		}
		return fmt.Errorf("getting column: %w", err)
	}
	board, err := s.db.GetBoardByID(ctx, col.BoardID)
	if err != nil {
		return fmt.Errorf("getting board for column: %w", err)
	}
	if err := requireRole(ctx, s.db, board.WorkspaceID, callerID, db.RoleAdmin); err != nil {
		return err
	}
	if err := s.db.DeleteColumn(ctx, id); err != nil {
		return fmt.Errorf("deleting column: %w", err)
	}
	return nil
}

func columnToModel(col db.KbColumn) model.Column {
	return model.Column{
		ID:         convert.UUIDToString(col.ID),
		Title:      col.Title,
		WipLimit:   col.WipLimit,
		Color:      convert.ColorToString(col.Color),
		IsTerminal: col.IsTerminal,
		Position:   col.Position,
		BoardID:    convert.UUIDToString(col.BoardID),
		CreatedAt:  convert.TimeToString(col.CreatedAt),
		UpdatedAt:  convert.NullableTimeToString(col.UpdatedAt),
	}
}
