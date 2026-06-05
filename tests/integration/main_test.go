package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"getkanbam.app/api/internal/db"
	"getkanbam.app/api/internal/handler"
	midd "getkanbam.app/api/internal/middleware"
	"getkanbam.app/api/internal/services"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testJWTSecret = "test-secret-for-integration-tests"

var srv *httptest.Server

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("kanbam_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "terminate postgres: %v\n", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection string: %v\n", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := applySchema(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "apply schema: %v\n", err)
		os.Exit(1)
	}

	srv = buildServer(pool)
	defer srv.Close()

	os.Exit(m.Run())
}

func applySchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS citext"); err != nil {
		return fmt.Errorf("create citext extension: %w", err)
	}
	schema, err := os.ReadFile("../../KanBam-Schema.sql")
	if err != nil {
		return fmt.Errorf("read schema file: %w", err)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		return fmt.Errorf("execute schema: %w", err)
	}
	return nil
}

func buildServer(pool *pgxpool.Pool) *httptest.Server {
	queries := db.New(pool)

	authSvc := services.NewAuthService(testJWTSecret, queries)
	workspaceSvc := services.NewWorkspaceService(queries)
	boardSvc := services.NewBoardService(queries)
	columnSvc := services.NewColumnService(queries)
	cardSvc := services.NewCardService(queries)
	commentSvc := services.NewCommentService(queries)
	tagSvc := services.NewTagService(queries)
	activitySvc := services.NewActivityService(queries)

	authH := handler.NewAuthHandler(&authSvc)
	workspaceH := handler.NewWorkspaceHandler(&workspaceSvc)
	boardH := handler.NewBoardHandler(&boardSvc, &columnSvc, &cardSvc)
	columnH := handler.NewColumnHandler(&columnSvc)
	cardH := handler.NewCardHandler(&cardSvc, &commentSvc, &boardSvc)
	commentH := handler.NewCommentHandler(&commentSvc)
	tagH := handler.NewTagHandler(&tagSvc)
	activityH := handler.NewActivityHandler(&activitySvc)

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Post("/auth/register", authH.Register)
	r.Post("/auth/login", authH.Login)
	r.With(midd.JWTBearerMiddleware(testJWTSecret)).Post("/auth/updatePassword", authH.UpdatePassword)

	r.Group(func(r chi.Router) {
		r.Use(midd.JWTBearerMiddleware(testJWTSecret))

		r.Get("/workspaces", workspaceH.List)
		r.Post("/workspaces/create", workspaceH.Create)
		r.Get("/workspaces/{workspaceID}", workspaceH.Get)
		r.Patch("/workspaces/{workspaceID}", workspaceH.Update)
		r.Delete("/workspaces/{workspaceID}", workspaceH.Delete)

		r.Get("/workspaces/{workspaceID}/members", workspaceH.ListMembers)
		r.Post("/workspaces/{workspaceID}/members", workspaceH.AddMember)
		r.Get("/workspaces/{workspaceID}/members/{userID}", workspaceH.GetMember)
		r.Patch("/workspaces/{workspaceID}/members/{userID}", workspaceH.UpdateMemberRole)
		r.Delete("/workspaces/{workspaceID}/members/{userID}", workspaceH.RemoveMember)

		r.Get("/workspaces/{workspaceID}/tags", workspaceH.ListTags)
		r.Post("/workspaces/{workspaceID}/tags/create", workspaceH.CreateTag)

		r.Get("/workspaces/{workspaceID}/activity", workspaceH.GetActivity)

		r.Get("/workspaces/{workspaceID}/boards", boardH.List)
		r.Post("/workspaces/{workspaceID}/boards/create", boardH.Create)

		r.Get("/boards/{boardID}", boardH.Get)
		r.Patch("/boards/{boardID}", boardH.Update)
		r.Delete("/boards/{boardID}", boardH.Delete)

		r.Get("/boards/{boardID}/columns", boardH.ListColumns)
		r.Post("/boards/{boardID}/columns/create", boardH.CreateColumn)

		r.Get("/boards/{boardID}/cards", boardH.ListCards)
		r.Post("/boards/{boardID}/cards/create", boardH.CreateCard)

		r.Get("/boards/{boardID}/activity", boardH.GetActivity)

		r.Get("/columns/{columnID}", columnH.Get)
		r.Patch("/columns/{columnID}", columnH.Update)
		r.Delete("/columns/{columnID}", columnH.Delete)

		r.Get("/cards/{cardID}", cardH.Get)
		r.Patch("/cards/{cardID}", cardH.Update)
		r.Delete("/cards/{cardID}", cardH.Delete)

		r.Get("/cards/{cardID}/comments", cardH.ListComments)
		r.Post("/cards/{cardID}/comments/create", cardH.CreateComment)

		r.Get("/cards/{cardID}/activity", cardH.GetActivity)

		r.Get("/comments/{commentID}", commentH.Get)
		r.Patch("/comments/{commentID}", commentH.Update)
		r.Delete("/comments/{commentID}", commentH.Delete)

		r.Get("/tags/{tagID}", tagH.Get)
		r.Patch("/tags/{tagID}", tagH.Update)
		r.Delete("/tags/{tagID}", tagH.Delete)

		r.Get("/activity/{activityID}", activityH.Get)
	})

	return httptest.NewServer(r)
}
