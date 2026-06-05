package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"getkanbam.app/api/internal/db"
	"getkanbam.app/api/internal/handler"
	midd "getkanbam.app/api/internal/middleware"
	"getkanbam.app/api/internal/services"
	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type config struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	JWTSecret   string `env:"JWT_SECRET,required"`
	Port        int    `env:"PORT" envDefault:"8080"`
}

func main() {
	_ = godotenv.Load()

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	migrateURL := strings.NewReplacer(
		"postgresql://", "pgx5://",
		"postgres://", "pgx5://",
	).Replace(cfg.DatabaseURL)
	m, err := migrate.New("file://migrations", migrateURL)
	if err != nil {
		log.Fatalf("migrations init: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrations up: %v", err)
	}

	queries := db.New(pool)

	// Services
	authSvc := services.NewAuthService(cfg.JWTSecret, queries)
	workspaceSvc := services.NewWorkspaceService(queries)
	boardSvc := services.NewBoardService(queries)
	columnSvc := services.NewColumnService(queries)
	cardSvc := services.NewCardService(queries)
	commentSvc := services.NewCommentService(queries)
	tagSvc := services.NewTagService(queries)
	activitySvc := services.NewActivityService(queries)

	// Handlers
	authHandler := handler.NewAuthHandler(&authSvc)
	workspaceHandler := handler.NewWorkspaceHandler(&workspaceSvc)
	boardHandler := handler.NewBoardHandler(&boardSvc, &columnSvc, &cardSvc)
	columnHandler := handler.NewColumnHandler(&columnSvc)
	cardHandler := handler.NewCardHandler(&cardSvc, &commentSvc, &boardSvc)
	commentHandler := handler.NewCommentHandler(&commentSvc)
	tagHandler := handler.NewTagHandler(&tagSvc)
	activityHandler := handler.NewActivityHandler(&activitySvc)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Auth (no JWT required except updatePassword)
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.With(midd.JWTBearerMiddleware(cfg.JWTSecret)).Post("/auth/updatePassword", authHandler.UpdatePassword)

	// All routes below require JWT
	r.Group(func(r chi.Router) {
		r.Use(midd.JWTBearerMiddleware(cfg.JWTSecret))

		// Workspaces
		r.Get("/workspaces", workspaceHandler.List)
		r.Post("/workspaces/create", workspaceHandler.Create)
		r.Get("/workspaces/{workspaceID}", workspaceHandler.Get)
		r.Patch("/workspaces/{workspaceID}", workspaceHandler.Update)
		r.Delete("/workspaces/{workspaceID}", workspaceHandler.Delete)

		// Workspace members
		r.Get("/workspaces/{workspaceID}/members", workspaceHandler.ListMembers)
		r.Post("/workspaces/{workspaceID}/members", workspaceHandler.AddMember)
		r.Get("/workspaces/{workspaceID}/members/{userID}", workspaceHandler.GetMember)
		r.Patch("/workspaces/{workspaceID}/members/{userID}", workspaceHandler.UpdateMemberRole)
		r.Delete("/workspaces/{workspaceID}/members/{userID}", workspaceHandler.RemoveMember)

		// Workspace tags
		r.Get("/workspaces/{workspaceID}/tags", workspaceHandler.ListTags)
		r.Post("/workspaces/{workspaceID}/tags/create", workspaceHandler.CreateTag)

		// Workspace activity
		r.Get("/workspaces/{workspaceID}/activity", workspaceHandler.GetActivity)

		// Boards (workspace-scoped creation/listing)
		r.Get("/workspaces/{workspaceID}/boards", boardHandler.List)
		r.Post("/workspaces/{workspaceID}/boards/create", boardHandler.Create)

		// Boards (direct access)
		r.Get("/boards/{boardID}", boardHandler.Get)
		r.Patch("/boards/{boardID}", boardHandler.Update)
		r.Delete("/boards/{boardID}", boardHandler.Delete)

		// Columns (board-scoped)
		r.Get("/boards/{boardID}/columns", boardHandler.ListColumns)
		r.Post("/boards/{boardID}/columns/create", boardHandler.CreateColumn)

		// Cards (board-scoped)
		r.Get("/boards/{boardID}/cards", boardHandler.ListCards)
		r.Post("/boards/{boardID}/cards/create", boardHandler.CreateCard)

		// Board activity
		r.Get("/boards/{boardID}/activity", boardHandler.GetActivity)

		// Columns (direct access)
		r.Get("/columns/{columnID}", columnHandler.Get)
		r.Patch("/columns/{columnID}", columnHandler.Update)
		r.Delete("/columns/{columnID}", columnHandler.Delete)

		// Cards (direct access)
		r.Get("/cards/{cardID}", cardHandler.Get)
		r.Patch("/cards/{cardID}", cardHandler.Update)
		r.Delete("/cards/{cardID}", cardHandler.Delete)

		// Comments (card-scoped)
		r.Get("/cards/{cardID}/comments", cardHandler.ListComments)
		r.Post("/cards/{cardID}/comments/create", cardHandler.CreateComment)

		// Card activity
		r.Get("/cards/{cardID}/activity", cardHandler.GetActivity)

		// Comments (direct access)
		r.Get("/comments/{commentID}", commentHandler.Get)
		r.Patch("/comments/{commentID}", commentHandler.Update)
		r.Delete("/comments/{commentID}", commentHandler.Delete)

		// Tags (direct access)
		r.Get("/tags/{tagID}", tagHandler.Get)
		r.Patch("/tags/{tagID}", tagHandler.Update)
		r.Delete("/tags/{tagID}", tagHandler.Delete)

		// Activity (direct access)
		r.Get("/activity/{activityID}", activityHandler.Get)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
