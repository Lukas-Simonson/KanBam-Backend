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
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	boards, err := h.boards.ListBoards(r.Context(), callerID, wsID)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"boards": boards})
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAndValidate[model.BoardCreation](w, r)
	if !ok {
		return
	}
	board, err := h.boards.CreateBoard(r.Context(), callerID, wsID, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, board)
}

func (h *BoardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	board, err := h.boards.GetBoard(r.Context(), callerID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, board)
}

func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAndValidate[model.BoardUpdate](w, r)
	if !ok {
		return
	}
	board, err := h.boards.UpdateBoard(r.Context(), callerID, id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, board)
}

func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	if err := h.boards.DeleteBoard(r.Context(), callerID, id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Columns (board-scoped) ---

func (h *BoardHandler) ListColumns(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	cols, err := h.columns.ListColumnsWithCards(r.Context(), callerID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"columns": cols})
}

func (h *BoardHandler) CreateColumn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAndValidate[model.ColumnCreation](w, r)
	if !ok {
		return
	}
	col, err := h.columns.CreateColumn(r.Context(), callerID, id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, col)
}

// --- Cards (board-scoped) ---

func (h *BoardHandler) ListCards(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "boardID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	cards, err := h.cards.ListCardsWithTags(r.Context(), callerID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"cards": cards})
}

func (h *BoardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	bID := chi.URLParam(r, "boardID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}

	board, err := h.boards.GetBoard(r.Context(), callerID, bID)
	if err != nil {
		handleError(w, err)
		return
	}

	req, ok := decodeAndValidate[model.CardCreation](w, r)
	if !ok {
		return
	}

	card, err := h.cards.CreateCard(r.Context(), callerID, bID, board.WorkspaceID, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, card)
}

// --- Activity (board-scoped) ---

func (h *BoardHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	bID := chi.URLParam(r, "boardID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	board, err := h.boards.GetBoard(r.Context(), callerID, bID)
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
	activities, err := h.boards.GetActivity(r.Context(), callerID, bID, board.WorkspaceID, filters)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"activities": activities})
}
