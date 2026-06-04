package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"getkanbam.app/api/internal/apierr"
	"getkanbam.app/api/internal/db"
	"getkanbam.app/api/internal/jwtutil"
	"getkanbam.app/api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	secret string
	db     *db.Queries
}

func NewAuthService(jwtSecret string, db *db.Queries) AuthService {
	return AuthService{
		secret: jwtSecret,
		db:     db,
	}
}

func (auth *AuthService) Register(ctx context.Context, registration model.Registration) (model.UserToken, error) {
	_, err := auth.db.GetUserByEmail(ctx, registration.Email)
	if err == nil {
		return model.UserToken{}, apierr.EmailTaken()
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return model.UserToken{}, fmt.Errorf("checking email: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.UserToken{}, err
	}

	user, err := auth.db.CreateUser(ctx, db.CreateUserParams{
		ID:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:         registration.Name,
		Email:        registration.Email,
		PasswordHash: string(hash), // not sure if this conversion from []byte to string works.
	})
	if err != nil {
		return model.UserToken{}, err
	}

	token, err := jwtutil.Sign(user.ID.String(), auth.secret)
	if err != nil {
		return model.UserToken{}, err
	}

	return model.UserToken{
		Token: token,
		User: model.User{
			ID:        user.ID.String(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		},
	}, nil
}

func (auth *AuthService) Login(ctx context.Context, email string, password string) (model.UserToken, error) {
	user, err := auth.db.GetUserByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserToken{}, apierr.InvalidCredentials()
		}
		return model.UserToken{}, err
	}

	// check if the password matches
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return model.UserToken{}, apierr.InvalidCredentials()
	}

	token, err := jwtutil.Sign(user.ID.String(), auth.secret)
	if err != nil {
		return model.UserToken{}, err
	}

	return model.UserToken{
		Token: token,
		User: model.User{
			ID:        user.ID.String(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		},
	}, nil
}

func (auth *AuthService) UpdatePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error {
	id := pgtype.UUID{}
	id.Scan(userID)

	user, err := auth.db.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return apierr.InvalidPassword()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = auth.db.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: string(hash),
	})
	if err != nil {
		return err
	}

	return nil
}
