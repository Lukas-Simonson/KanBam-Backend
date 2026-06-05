package integration

import (
	"net/http"
	"testing"
)

// --- Create ---

func TestCreateWorkspace_Success(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPost, "/workspaces/create", map[string]any{
		"title":       "My Workspace",
		"description": "A test workspace",
	}, user.Token)
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "id")
	mustHaveField(t, out, "title")
}

func TestCreateWorkspace_MissingTitle(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPost, "/workspaces/create", map[string]any{
		"description": "No title here",
	}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestCreateWorkspace_NoAuth(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/workspaces/create", map[string]any{
		"title": "Sneaky",
	}, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- List ---

func TestListWorkspaces_Success(t *testing.T) {
	user := registerUser(t)
	createWorkspace(t, user.Token)
	createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Workspaces []map[string]any `json:"workspaces"`
	}
	decodeBody(t, resp, &out)
	if len(out.Workspaces) < 2 {
		t.Errorf("expected at least 2 workspaces, got %d", len(out.Workspaces))
	}
}

func TestListWorkspaces_NoAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/workspaces", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Get ---

func TestGetWorkspace_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID, nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["id"] != wsID {
		t.Errorf("id mismatch: got %v, want %v", out["id"], wsID)
	}
}

func TestGetWorkspace_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/workspaces/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "WORKSPACE_NOT_FOUND")
}

func TestGetWorkspace_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID, nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Update ---

func TestUpdateWorkspace_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodPatch, "/workspaces/"+wsID, map[string]any{
		"title": "Updated Title",
	}, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["title"] != "Updated Title" {
		t.Errorf("title mismatch: got %v", out["title"])
	}
}

func TestUpdateWorkspace_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = owner

	resp := doRequest(t, http.MethodPatch, "/workspaces/"+wsID, map[string]any{
		"title": "Viewer Update",
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestUpdateWorkspace_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPatch, "/workspaces/"+nonexistentUUID, map[string]any{
		"title": "Ghost",
	}, user.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Delete ---

func TestDeleteWorkspace_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodDelete, "/workspaces/"+wsID, nil, user.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)

	// Should be gone now
	getResp := doRequest(t, http.MethodGet, "/workspaces/"+wsID, nil, user.Token)
	assertStatus(t, getResp, http.StatusNotFound)
	drainBody(getResp)
}

func TestDeleteWorkspace_InsufficientRole(t *testing.T) {
	owner, admin, wsID := setupWorkspaceWithMember(t, "admin")
	_ = owner

	resp := doRequest(t, http.MethodDelete, "/workspaces/"+wsID, nil, admin.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestDeleteWorkspace_NoAuth(t *testing.T) {
	owner := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodDelete, "/workspaces/"+wsID, nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Members ---

func TestListMembers_Success(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = viewer

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/members", nil, owner.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Members []map[string]any `json:"members"`
	}
	decodeBody(t, resp, &out)
	if len(out.Members) < 2 {
		t.Errorf("expected at least 2 members, got %d", len(out.Members))
	}
}

func TestListMembers_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/members", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestAddMember_Success(t *testing.T) {
	owner := registerUser(t)
	newMember := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/members", map[string]any{
		"userID": newMember.ID,
		"role":   "viewer",
	}, owner.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)
}

func TestAddMember_AlreadyMember(t *testing.T) {
	owner, member, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = owner

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/members", map[string]any{
		"userID": member.ID,
		"role":   "viewer",
	}, owner.Token)
	assertStatus(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "ALREADY_IN_WORKSPACE")
}

func TestAddMember_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	extra := registerUser(t)
	_ = owner

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/members", map[string]any{
		"userID": extra.ID,
		"role":   "viewer",
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestAddMember_InvalidRole(t *testing.T) {
	owner := registerUser(t)
	newMember := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/members", map[string]any{
		"userID": newMember.ID,
		"role":   "superadmin",
	}, owner.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestGetMember_Success(t *testing.T) {
	owner, member, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = owner

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/members/"+member.ID, nil, owner.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "id")
}

func TestGetMember_NotFound(t *testing.T) {
	owner := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	outsider := registerUser(t)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/members/"+outsider.ID, nil, owner.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestUpdateMemberRole_Success(t *testing.T) {
	owner, member, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = owner

	resp := doRequest(t, http.MethodPatch, "/workspaces/"+wsID+"/members/"+member.ID, map[string]any{
		"role": "admin",
	}, owner.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)
}

func TestUpdateMemberRole_InsufficientRole(t *testing.T) {
	owner, admin, wsID := setupWorkspaceWithMember(t, "admin")
	viewer := registerUser(t)
	addMember(t, owner.Token, wsID, viewer.ID, "viewer")

	resp := doRequest(t, http.MethodPatch, "/workspaces/"+wsID+"/members/"+viewer.ID, map[string]any{
		"role": "admin",
	}, admin.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestRemoveMember_Success(t *testing.T) {
	owner, member, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = owner

	resp := doRequest(t, http.MethodDelete, "/workspaces/"+wsID+"/members/"+member.ID, nil, owner.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)
}

func TestRemoveMember_SelfRemoval(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = owner

	resp := doRequest(t, http.MethodDelete, "/workspaces/"+wsID+"/members/"+viewer.ID, nil, viewer.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)
}

func TestRemoveMember_InsufficientRole(t *testing.T) {
	owner, viewerA, wsID := setupWorkspaceWithMember(t, "viewer")
	viewerB := registerUser(t)
	addMember(t, owner.Token, wsID, viewerB.ID, "viewer")

	resp := doRequest(t, http.MethodDelete, "/workspaces/"+wsID+"/members/"+viewerB.ID, nil, viewerA.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}
