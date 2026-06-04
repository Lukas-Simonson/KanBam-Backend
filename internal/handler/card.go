package handler

import (
	"encoding/json"
	"net/http"

	"getkanbam.app/api/internal/middleware"
	"getkanbam.app/api/internal/model"
	"getkanbam.app/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type CardHandler struct {
	cards    *services.CardService
	comments *services.CommentService
	boards   *services.BoardService
}

func NewCardHandler(c *services.CardService, com *services.CommentService, b *services.BoardService) CardHandler {
	return CardHandler{cards: c, comments: com, boards: b}
}

func (h *CardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cardID")
	card, err := h.cards.GetCardWithTags(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, card)
}

func (h *CardHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cardID")
	var req model.CardUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}
	card, err := h.cards.UpdateCard(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, card)
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cardID")
	if err := h.cards.DeleteCard(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Comments (card-scoped) ---

func (h *CardHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cardID")
	comments, err := h.comments.ListComments(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"comments": comments})
}

func (h *CardHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		return
	}
	var req model.CommentCreation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}
	comment, err := h.comments.CreateComment(r.Context(), cardID, userID, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, comment)
}

// --- Activity (card-scoped) ---

func (h *CardHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cardID")
	q := r.URL.Query()
	filters := services.ActivityFilters{
		UserID: parseQueryUUID(q.Get("userID")),
		After:  parseQueryTime(q.Get("after")),
		Before: parseQueryTime(q.Get("before")),
	}
	activities, err := h.cards.GetActivity(r.Context(), id, filters)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"activities": activities})
}
