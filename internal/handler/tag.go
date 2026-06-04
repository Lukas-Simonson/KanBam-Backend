package handler

import (
	"net/http"

	"getkanbam.app/api/internal/model"
	"getkanbam.app/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type TagHandler struct {
	tags *services.TagService
}

func NewTagHandler(t *services.TagService) TagHandler {
	return TagHandler{tags: t}
}

func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tagID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	tag, err := h.tags.GetTag(r.Context(), callerID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, tag)
}

func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tagID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAndValidate[model.TagUpdate](w, r)
	if !ok {
		return
	}
	tag, err := h.tags.UpdateTag(r.Context(), callerID, id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, tag)
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "tagID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	if err := h.tags.DeleteTag(r.Context(), callerID, id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
