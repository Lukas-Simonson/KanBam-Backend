package handler

import (
	"net/http"

	"getkanbam.app/api/internal/model"
	"getkanbam.app/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type CommentHandler struct {
	comments *services.CommentService
}

func NewCommentHandler(c *services.CommentService) CommentHandler {
	return CommentHandler{comments: c}
}

func (h *CommentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "commentID")
	comment, err := h.comments.GetComment(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "commentID")
	req, ok := decodeAndValidate[model.CommentUpdate](w, r)
	if !ok {
		return
	}
	comment, err := h.comments.UpdateComment(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "commentID")
	if err := h.comments.DeleteComment(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
