package integration

import (
	"net/http"
	"testing"
)

// --- Create Tag (workspace-scoped) ---

func TestCreateTag_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/tags/create", map[string]any{
		"name":  "Bug",
		"color": "#FF0000",
	}, user.Token)
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "id")
	if out["name"] != "Bug" {
		t.Errorf("name mismatch: got %v", out["name"])
	}
}

func TestCreateTag_DuplicateName(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/tags/create", map[string]any{
		"name":  "Duplicate",
		"color": "#FF0000",
	}, user.Token).Body.Close()

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/tags/create", map[string]any{
		"name":  "Duplicate",
		"color": "#00FF00",
	}, user.Token)
	assertStatus(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "TAG_NAME_TAKEN")
}

func TestCreateTag_MissingName(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/tags/create", map[string]any{
		"color": "#FF0000",
	}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestCreateTag_InvalidColor(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/tags/create", map[string]any{
		"name":  "BadColor",
		"color": "red",
	}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestCreateTag_ViewerRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = owner

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/tags/create", map[string]any{
		"name":  "ViewerTag",
		"color": "#FF0000",
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestCreateTag_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/tags/create", map[string]any{
		"name":  "OutsiderTag",
		"color": "#FF0000",
	}, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- List Tags (workspace-scoped) ---

func TestListTags_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	createTag(t, user.Token, wsID)
	createTag(t, user.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/tags", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Tags []map[string]any `json:"tags"`
	}
	decodeBody(t, resp, &out)
	if len(out.Tags) < 2 {
		t.Errorf("expected at least 2 tags, got %d", len(out.Tags))
	}
}

func TestListTags_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/tags", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Get Tag ---

func TestGetTag_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	tagID := createTag(t, user.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/tags/"+tagID, nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["id"] != tagID {
		t.Errorf("id mismatch: got %v, want %v", out["id"], tagID)
	}
}

func TestGetTag_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/tags/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "TAG_NOT_FOUND")
}

func TestGetTag_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	tagID := createTag(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/tags/"+tagID, nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Update Tag ---

func TestUpdateTag_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	tagID := createTag(t, user.Token, wsID)

	resp := doRequest(t, http.MethodPatch, "/tags/"+tagID, map[string]any{
		"name":  "UpdatedTag",
		"color": "#0000FF",
	}, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["name"] != "UpdatedTag" {
		t.Errorf("name mismatch: got %v", out["name"])
	}
}

func TestUpdateTag_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPatch, "/tags/"+nonexistentUUID, map[string]any{
		"name": "Ghost",
	}, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "TAG_NOT_FOUND")
}

func TestUpdateTag_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	tagID := createTag(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodPatch, "/tags/"+tagID, map[string]any{
		"name": "ViewerUpdate",
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Delete Tag ---

func TestDeleteTag_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	tagID := createTag(t, user.Token, wsID)

	resp := doRequest(t, http.MethodDelete, "/tags/"+tagID, nil, user.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)

	getResp := doRequest(t, http.MethodGet, "/tags/"+tagID, nil, user.Token)
	assertStatus(t, getResp, http.StatusNotFound)
	drainBody(getResp)
}

func TestDeleteTag_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodDelete, "/tags/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "TAG_NOT_FOUND")
}

func TestDeleteTag_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	tagID := createTag(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodDelete, "/tags/"+tagID, nil, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}
