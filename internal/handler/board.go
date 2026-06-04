package handler

import (
	"net/http"

	"getkanbam.app/api/internal/model"
	"getkanbam.app/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type BoardHandler struct {
	boards  *services.BoardService
	columns *services.ColumnService
	cards   *services.CardService
}

func NewBoardHandler(b *services.BoardService, col *services.ColumnService, c *services.CardService) BoardHandler {
	return BoardHandler{boards: b, columns: col, cards: c}
}

func (h *BoardHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	boards, err := h.boards.ListBoards(r.Context(), wsID)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"boards": boards})
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	req, ok := decodeAndValidate[model.BoardCreation](w, r)
	if !ok {
		return
	}
	board, err := h.boards.CreateBoard(r.Context(), wsID, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, board)
}

func (h *BoardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	board, err := h.boards.GetBoard(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, board)
}

func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	req, ok := decodeAndValidate[model.BoardUpdate](w, r)
	if !ok {
		return
	}
	board, err := h.boards.UpdateBoard(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, board)
}

func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	if err := h.boards.DeleteBoard(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Columns (board-scoped) ---

func (h *BoardHandler) ListColumns(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	cols, err := h.columns.ListColumnsWithCards(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"columns": cols})
}

func (h *BoardHandler) CreateColumn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	req, ok := decodeAndValidate[model.ColumnCreation](w, r)
	if !ok {
		return
	}
	col, err := h.columns.CreateColumn(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, col)
}

// --- Cards (board-scoped) ---

func (h *BoardHandler) ListCards(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	cards, err := h.cards.ListCardsWithTags(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"cards": cards})
}

func (h *BoardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	bID := chi.URLParam(r, "boardID")
	wsID := chi.URLParam(r, "workspaceID")

	board, err := h.boards.GetBoard(r.Context(), bID)
	if err != nil {
		handleError(w, err)
		return
	}
	if wsID == "" {
		wsID = board.WorkspaceID
	}

	req, ok := decodeAndValidate[model.CardCreation](w, r)
	if !ok {
		return
	}

	card, err := h.cards.CreateCard(r.Context(), bID, wsID, "", req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, card)
}

// --- Activity (board-scoped) ---

func (h *BoardHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	bID := chi.URLParam(r, "boardID")
	board, err := h.boards.GetBoard(r.Context(), bID)
	if err != nil {
		handleError(w, err)
		return
	}
	q := r.URL.Query()
	filters := services.ActivityFilters{
		CardID: parseQueryUUID(q.Get("cardID")),
		UserID: parseQueryUUID(q.Get("userID")),
		After:  parseQueryTime(q.Get("after")),
		Before: parseQueryTime(q.Get("before")),
	}
	activities, err := h.boards.GetActivity(r.Context(), bID, board.WorkspaceID, filters)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"activities": activities})
}
