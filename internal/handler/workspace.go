package handler

import (
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
	req, ok := decodeAndValidate[model.WorkspaceCreation](w, r)
	if !ok {
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
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	ws, err := h.workspaces.GetWorkspace(r.Context(), callerID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAndValidate[model.WorkspaceUpdate](w, r)
	if !ok {
		return
	}
	ws, err := h.workspaces.UpdateWorkspace(r.Context(), callerID, id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	if err := h.workspaces.DeleteWorkspace(r.Context(), callerID, id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Members ---

func (h *WorkspaceHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	members, err := h.workspaces.ListMembers(r.Context(), callerID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"members": members})
}

func (h *WorkspaceHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAndValidate[model.Membership](w, r)
	if !ok {
		return
	}
	if err := h.workspaces.AddMember(r.Context(), callerID, id, req); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	uID := chi.URLParam(r, "userID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	member, err := h.workspaces.GetMember(r.Context(), callerID, wsID, uID)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, member)
}

func (h *WorkspaceHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	uID := chi.URLParam(r, "userID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAndValidate[model.MembershipUpdate](w, r)
	if !ok {
		return
	}
	if err := h.workspaces.UpdateMemberRole(r.Context(), callerID, wsID, uID, req); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "workspaceID")
	uID := chi.URLParam(r, "userID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	if err := h.workspaces.RemoveMember(r.Context(), callerID, wsID, uID); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Tags ---

func (h *WorkspaceHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	tags, err := h.workspaces.ListTags(r.Context(), callerID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"tags": tags})
}

func (h *WorkspaceHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAndValidate[model.TagCreation](w, r)
	if !ok {
		return
	}
	tag, err := h.workspaces.CreateTag(r.Context(), callerID, id, req)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, tag)
}

// --- Activity ---

func (h *WorkspaceHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	callerID, ok := mustCallerID(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	filters := services.ActivityFilters{
		BoardID: parseQueryUUID(q.Get("boardID")),
		CardID:  parseQueryUUID(q.Get("cardID")),
		UserID:  parseQueryUUID(q.Get("userID")),
		After:   parseQueryTime(q.Get("after")),
		Before:  parseQueryTime(q.Get("before")),
	}
	activities, err := h.workspaces.GetActivity(r.Context(), callerID, id, filters)
	if err != nil {
		handleError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"activities": activities})
}
