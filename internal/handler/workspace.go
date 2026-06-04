package handler

import (
	"encoding/json"
	"net/http"

	"getkanbam.app/api/internal/middleware"
	"getkanbam.app/api/internal/model"
	"getkanbam.app/api/internal/services"
	"github.com/go-chi/chi/v5"
)

type WorkspaceHandler struct {
	workspaces *services.WorkspaceService
}

func NewWorkspaceHandler(ws *services.WorkspaceService) WorkspaceHandler {
	return WorkspaceHandler{workspaces: ws}
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		return
	}
	ws, err := h.workspaces.ListWorkspaces(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"workspaces": ws})
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		return
	}
	var req model.WorkspaceCreation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}
	ws, err := h.workspaces.CreateWorkspace(r.Context(), userID, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, ws)
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	ws, err := h.workspaces.GetWorkspace(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	var req model.WorkspaceUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}
	ws, err := h.workspaces.UpdateWorkspace(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	if err := h.workspaces.DeleteWorkspace(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Members ---

func (h *WorkspaceHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	members, err := h.workspaces.ListMembers(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"members": members})
}

func (h *WorkspaceHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	var req model.Membership
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}
	if err := h.workspaces.AddMember(r.Context(), id, req); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	uID := chi.URLParam(r, "userID")
	member, err := h.workspaces.GetMember(r.Context(), wsID, uID)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, member)
}

func (h *WorkspaceHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	uID := chi.URLParam(r, "userID")
	var req model.MembershipUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}
	if err := h.workspaces.UpdateMemberRole(r.Context(), wsID, uID, req); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	uID := chi.URLParam(r, "userID")
	if err := h.workspaces.RemoveMember(r.Context(), wsID, uID); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Tags ---

func (h *WorkspaceHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	tags, err := h.workspaces.ListTags(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"tags": tags})
}

func (h *WorkspaceHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	var req model.TagCreation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return
	}
	tag, err := h.workspaces.CreateTag(r.Context(), id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, tag)
}

// --- Activity ---

func (h *WorkspaceHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	q := r.URL.Query()
	filters := services.ActivityFilters{
		BoardID: parseQueryUUID(q.Get("boardID")),
		CardID:  parseQueryUUID(q.Get("cardID")),
		UserID:  parseQueryUUID(q.Get("userID")),
		After:   parseQueryTime(q.Get("after")),
		Before:  parseQueryTime(q.Get("before")),
	}
	activities, err := h.workspaces.GetActivity(r.Context(), id, filters)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"activities": activities})
}
