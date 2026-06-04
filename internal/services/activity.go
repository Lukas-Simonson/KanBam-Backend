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

type ActivityService struct {
	db *db.Queries
}

func NewActivityService(q *db.Queries) ActivityService {
	return ActivityService{db: q}
}

func (s *ActivityService) GetActivity(ctx context.Context, activityID string) (model.Activity, error) {
	id, err := convert.ParseUUID(activityID)
	if err != nil {
		return model.Activity{}, fmt.Errorf("parsing activity id: %w", err)
	}
	a, err := s.db.GetActivityByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Activity{}, apierr.ActivityNotFound()
		}
		return model.Activity{}, fmt.Errorf("getting activity: %w", err)
	}
	return activityToModel(a), nil
}
