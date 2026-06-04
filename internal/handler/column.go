package handler

import (
	"net/http"

	"getkanbam.app/api/internal/model"
	"getkanbam.app/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type ColumnHandler struct {
	columns *services.ColumnService
}

func NewColumnHandler(c *services.ColumnService) ColumnHandler {
	return ColumnHandler{columns: c}
}

func (h *ColumnHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "columnID")
	col, err := h.columns.GetColumn(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, col)
}

func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "columnID")
	req, ok := decodeAndValidate[model.ColumnUpdate](w, r)
	if !ok {
		return
	}
	col, err := h.columns.UpdateColumn(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, col)
}

func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "columnID")
	if err := h.columns.DeleteColumn(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
