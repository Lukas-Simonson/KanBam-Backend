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

type TagService struct {
	db *db.Queries
}

func NewTagService(q *db.Queries) TagService {
	return TagService{db: q}
}

func (s *TagService) GetTag(ctx context.Context, tagID string) (model.Tag, error) {
	id, err := convert.ParseUUID(tagID)
	if err != nil {
		return model.Tag{}, fmt.Errorf("parsing tag id: %w", err)
	}
	t, err := s.db.GetTagByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Tag{}, apierr.TagNotFound()
		}
		return model.Tag{}, fmt.Errorf("getting tag: %w", err)
	}
	return tagToModel(t), nil
}

func (s *TagService) UpdateTag(ctx context.Context, tagID string, req model.TagUpdate) (model.Tag, error) {
	id, err := convert.ParseUUID(tagID)
	if err != nil {
		return model.Tag{}, fmt.Errorf("parsing tag id: %w", err)
	}
	current, err := s.db.GetTagByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Tag{}, apierr.TagNotFound()
		}
		return model.Tag{}, fmt.Errorf("getting tag: %w", err)
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	var color interface{} = current.Color
	if req.Color != nil {
		color = *req.Color
	}

	t, err := s.db.UpdateTag(ctx, db.UpdateTagParams{
		ID:    id,
		Name:  name,
		Color: color,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return model.Tag{}, apierr.TagNameTaken()
		}
		return model.Tag{}, fmt.Errorf("updating tag: %w", err)
	}
	return tagToModel(t), nil
}

func (s *TagService) DeleteTag(ctx context.Context, tagID string) error {
	id, err := convert.ParseUUID(tagID)
	if err != nil {
		return fmt.Errorf("parsing tag id: %w", err)
	}
	if err := s.db.DeleteTag(ctx, id); err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	return nil
}
