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

type CardService struct {
	db *db.Queries
}

func NewCardService(q *db.Queries) CardService {
	return CardService{db: q}
}

func (s *CardService) ListCardsWithTags(ctx context.Context, boardID string) ([]model.CardWithTags, error) {
	bID, err := convert.ParseUUID(boardID)
	if err != nil {
		return nil, fmt.Errorf("parsing board id: %w", err)
	}
	cards, err := s.db.GetCardsByBoard(ctx, bID)
	if err != nil {
		return nil, fmt.Errorf("listing cards: %w", err)
	}
	result := make([]model.CardWithTags, len(cards))
	for i, c := range cards {
		tags, err := s.db.GetCardTags(ctx, c.ID)
		if err != nil {
			return nil, fmt.Errorf("getting card tags: %w", err)
		}
		result[i] = cardWithTagsToModel(c, tags)
	}
	return result, nil
}

func (s *CardService) GetCardWithTags(ctx context.Context, cardID string) (model.CardWithTags, error) {
	id, err := convert.ParseUUID(cardID)
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("parsing card id: %w", err)
	}
	c, err := s.db.GetCardByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CardWithTags{}, apierr.CardNotFound()
		}
		return model.CardWithTags{}, fmt.Errorf("getting card: %w", err)
	}
	tags, err := s.db.GetCardTags(ctx, id)
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("getting card tags: %w", err)
	}
	return cardWithTagsToModel(c, tags), nil
}

// CreateCard creates a card, increments the board sequence for the reference,
// and attaches any provided tag IDs.
func (s *CardService) CreateCard(ctx context.Context, boardID, workspaceID, userID string, req model.CardCreation) (model.CardWithTags, error) {
	bID, err := convert.ParseUUID(boardID)
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("parsing board id: %w", err)
	}
	wsID, err := convert.ParseUUID(workspaceID)
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("parsing workspace id: %w", err)
	}

	// Fetch board to build the reference (e.g. #PROJ-42).
	board, err := s.db.GetBoardByID(ctx, bID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CardWithTags{}, apierr.BoardNotFound()
		}
		return model.CardWithTags{}, fmt.Errorf("getting board: %w", err)
	}
	seq, err := s.db.IncrementBoardSequence(ctx, bID)
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("incrementing board sequence: %w", err)
	}
	reference := fmt.Sprintf("#%s-%d", board.Prefix, seq)

	columnID := nullableUUID(req.ColumnID)
	assigneeID := nullableUUID(req.UserID)

	card, err := s.db.CreateCard(ctx, db.CreateCardParams{
		ID:          convert.NewUUID(),
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
		Reference:   reference,
		WorkspaceID: wsID,
		BoardID:     bID,
		ColumnID:    columnID,
		UserID:      assigneeID,
	})
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("creating card: %w", err)
	}

	// Attach tags.
	for _, tagIDStr := range req.TagIDs {
		tagID, err := convert.ParseUUID(tagIDStr)
		if err != nil {
			return model.CardWithTags{}, fmt.Errorf("parsing tag id: %w", err)
		}
		if err := s.db.AddCardTag(ctx, db.AddCardTagParams{
			CardID: card.ID,
			TagID:  tagID,
		}); err != nil {
			return model.CardWithTags{}, fmt.Errorf("adding card tag: %w", err)
		}
	}

	tags, err := s.db.GetCardTags(ctx, card.ID)
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("getting card tags: %w", err)
	}
	return cardWithTagsToModel(card, tags), nil
}

// UpdateCard applies a partial update using read-then-write.
//
// NOTE: Setting column_id or user_id to NULL (unassigning) is not supported
// by the current COALESCE-based query. Add dedicated NullifyCardColumn /
// NullifyCardAssignee queries if that behaviour is required.
func (s *CardService) UpdateCard(ctx context.Context, cardID string, req model.CardUpdate) (model.CardWithTags, error) {
	id, err := convert.ParseUUID(cardID)
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("parsing card id: %w", err)
	}
	current, err := s.db.GetCardByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CardWithTags{}, apierr.CardNotFound()
		}
		return model.CardWithTags{}, fmt.Errorf("getting card: %w", err)
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
	position := current.Position
	if req.Position != nil {
		position = *req.Position
	}
	columnID := current.ColumnID
	if req.ColumnID != nil {
		columnID = nullableUUID(req.ColumnID)
	}
	assigneeID := current.UserID
	if req.UserID != nil {
		assigneeID = nullableUUID(req.UserID)
	}

	card, err := s.db.UpdateCard(ctx, db.UpdateCardParams{
		ID:          id,
		Title:       title,
		Description: description,
		IsArchived:  isArchived,
		Position:    position,
		ColumnID:    columnID,
		UserID:      assigneeID,
	})
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("updating card: %w", err)
	}

	tags, err := s.db.GetCardTags(ctx, id)
	if err != nil {
		return model.CardWithTags{}, fmt.Errorf("getting card tags: %w", err)
	}
	return cardWithTagsToModel(card, tags), nil
}

func (s *CardService) DeleteCard(ctx context.Context, cardID string) error {
	id, err := convert.ParseUUID(cardID)
	if err != nil {
		return fmt.Errorf("parsing card id: %w", err)
	}
	if err := s.db.DeleteCard(ctx, id); err != nil {
		return fmt.Errorf("deleting card: %w", err)
	}
	return nil
}

func (s *CardService) GetActivity(ctx context.Context, cardID string, filters ActivityFilters) ([]model.Activity, error) {
	id, err := convert.ParseUUID(cardID)
	if err != nil {
		return nil, fmt.Errorf("parsing card id: %w", err)
	}
	rows, err := s.db.GetCardActivity(ctx, db.GetCardActivityParams{
		CardID:  id,
		Column2: filters.UserID,
		Column3: filters.After,
		Column4: filters.Before,
	})
	if err != nil {
		return nil, fmt.Errorf("getting card activity: %w", err)
	}
	return activitiesToModel(rows), nil
}

func cardToModel(c db.Card) model.Card {
	return model.Card{
		ID:          convert.UUIDToString(c.ID),
		Title:       c.Title,
		Description: c.Description,
		IsArchived:  c.IsArchived,
		Position:    c.Position,
		Reference:   c.Reference,
		BoardID:     convert.UUIDToString(c.BoardID),
		ColumnID:    convert.NullableUUIDToString(c.ColumnID),
		UserID:      convert.NullableUUIDToString(c.UserID),
		CreatedAt:   convert.TimeToString(c.CreatedAt),
		UpdatedAt:   convert.NullableTimeToString(c.UpdatedAt),
	}
}

func cardWithTagsToModel(c db.Card, tags []db.Tag) model.CardWithTags {
	modelTags := make([]model.Tag, len(tags))
	for i, t := range tags {
		modelTags[i] = tagToModel(t)
	}
	return model.CardWithTags{
		Card: cardToModel(c),
		Tags: modelTags,
	}
}
