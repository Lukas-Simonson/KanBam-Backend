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

type CommentService struct {
	db *db.Queries
}

func NewCommentService(q *db.Queries) CommentService {
	return CommentService{db: q}
}

func (s *CommentService) ListComments(ctx context.Context, cardID string) ([]model.Comment, error) {
	id, err := convert.ParseUUID(cardID)
	if err != nil {
		return nil, fmt.Errorf("parsing card id: %w", err)
	}
	rows, err := s.db.GetCommentsByCard(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}
	result := make([]model.Comment, len(rows))
	for i, c := range rows {
		result[i] = commentToModel(c)
	}
	return result, nil
}

func (s *CommentService) GetComment(ctx context.Context, commentID string) (model.Comment, error) {
	id, err := convert.ParseUUID(commentID)
	if err != nil {
		return model.Comment{}, fmt.Errorf("parsing comment id: %w", err)
	}
	c, err := s.db.GetCommentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Comment{}, apierr.CommentNotFound()
		}
		return model.Comment{}, fmt.Errorf("getting comment: %w", err)
	}
	return commentToModel(c), nil
}

func (s *CommentService) CreateComment(ctx context.Context, cardID, userID string, req model.CommentCreation) (model.Comment, error) {
	cID, err := convert.ParseUUID(cardID)
	if err != nil {
		return model.Comment{}, fmt.Errorf("parsing card id: %w", err)
	}
	uID, err := convert.ParseUUID(userID)
	if err != nil {
		return model.Comment{}, fmt.Errorf("parsing user id: %w", err)
	}
	c, err := s.db.CreateComment(ctx, db.CreateCommentParams{
		ID:     convert.NewUUID(),
		Body:   req.Body,
		CardID: cID,
		UserID: uID,
	})
	if err != nil {
		return model.Comment{}, fmt.Errorf("creating comment: %w", err)
	}
	return commentToModel(c), nil
}

func (s *CommentService) UpdateComment(ctx context.Context, commentID string, req model.CommentUpdate) (model.Comment, error) {
	id, err := convert.ParseUUID(commentID)
	if err != nil {
		return model.Comment{}, fmt.Errorf("parsing comment id: %w", err)
	}
	c, err := s.db.UpdateComment(ctx, db.UpdateCommentParams{
		ID:   id,
		Body: req.Body,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Comment{}, apierr.CommentNotFound()
		}
		return model.Comment{}, fmt.Errorf("updating comment: %w", err)
	}
	return commentToModel(c), nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID string) error {
	id, err := convert.ParseUUID(commentID)
	if err != nil {
		return fmt.Errorf("parsing comment id: %w", err)
	}
	if err := s.db.DeleteComment(ctx, id); err != nil {
		return fmt.Errorf("deleting comment: %w", err)
	}
	return nil
}

func commentToModel(c db.Comment) model.Comment {
	return model.Comment{
		ID:        convert.UUIDToString(c.ID),
		Body:      c.Body,
		CardID:    convert.UUIDToString(c.CardID),
		UserID:    convert.NullableUUIDToString(c.UserID),
		CreatedAt: convert.TimeToString(c.CreatedAt),
		UpdatedAt: convert.NullableTimeToString(c.UpdatedAt),
	}
}
