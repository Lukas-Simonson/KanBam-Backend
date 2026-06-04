package handler

import (
	"net/http"

	"getkanbam.app/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type ActivityHandler struct {
	activity *services.ActivityService
}

func NewActivityHandler(a *services.ActivityService) ActivityHandler {
	return ActivityHandler{activity: a}
}

func (h *ActivityHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "activityID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	activity, err := h.activity.GetActivity(r.Context(), callerID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, activity)
}
