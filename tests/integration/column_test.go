package integration

import (
	"net/http"
	"testing"
)

// --- Create Column (board-scoped) ---

func TestCreateColumn_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/columns/create", map[string]any{
		"title":    "To Do",
		"position": 0,
	}, user.Token)
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "id")
	mustHaveField(t, out, "title")
}

func TestCreateColumn_WithColor(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/columns/create", map[string]any{
		"title":    "In Progress",
		"position": 1,
		"color":    "#00FF00",
	}, user.Token)
	assertStatus(t, resp, http.StatusCreated)
	drainBody(resp)
}

func TestCreateColumn_MissingTitle(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/columns/create", map[string]any{
		"position": 0,
	}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestCreateColumn_InvalidColor(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/columns/create", map[string]any{
		"title":    "Bad Color",
		"position": 0,
		"color":    "not-a-color",
	}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestCreateColumn_ViewerRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	bID := createBoard(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/columns/create", map[string]any{
		"title":    "Blocked",
		"position": 0,
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestCreateColumn_BoardNotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPost, "/boards/"+nonexistentUUID+"/columns/create", map[string]any{
		"title":    "Ghost Column",
		"position": 0,
	}, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "BOARD_NOT_FOUND")
}

// --- List Columns (board-scoped) ---

func TestListColumns_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	createColumn(t, user.Token, bID)
	createColumn(t, user.Token, bID)

	resp := doRequest(t, http.MethodGet, "/boards/"+bID+"/columns", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Columns []map[string]any `json:"columns"`
	}
	decodeBody(t, resp, &out)
	if len(out.Columns) < 2 {
		t.Errorf("expected at least 2 columns, got %d", len(out.Columns))
	}
}

func TestListColumns_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/boards/"+bID+"/columns", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Get Column ---

func TestGetColumn_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	colID := createColumn(t, user.Token, bID)

	resp := doRequest(t, http.MethodGet, "/columns/"+colID, nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["id"] != colID {
		t.Errorf("id mismatch: got %v, want %v", out["id"], colID)
	}
}

func TestGetColumn_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/columns/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "COLUMN_NOT_FOUND")
}

func TestGetColumn_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)
	colID := createColumn(t, owner.Token, bID)

	resp := doRequest(t, http.MethodGet, "/columns/"+colID, nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Update Column ---

func TestUpdateColumn_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	colID := createColumn(t, user.Token, bID)

	resp := doRequest(t, http.MethodPatch, "/columns/"+colID, map[string]any{
		"title": "Updated Column",
	}, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["title"] != "Updated Column" {
		t.Errorf("title mismatch: got %v", out["title"])
	}
}

func TestUpdateColumn_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPatch, "/columns/"+nonexistentUUID, map[string]any{
		"title": "Ghost",
	}, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "COLUMN_NOT_FOUND")
}

func TestUpdateColumn_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	bID := createBoard(t, owner.Token, wsID)
	colID := createColumn(t, owner.Token, bID)

	resp := doRequest(t, http.MethodPatch, "/columns/"+colID, map[string]any{
		"title": "Viewer Update",
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Delete Column ---

func TestDeleteColumn_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	colID := createColumn(t, user.Token, bID)

	resp := doRequest(t, http.MethodDelete, "/columns/"+colID, nil, user.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)

	getResp := doRequest(t, http.MethodGet, "/columns/"+colID, nil, user.Token)
	assertStatus(t, getResp, http.StatusNotFound)
	drainBody(getResp)
}

func TestDeleteColumn_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodDelete, "/columns/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "COLUMN_NOT_FOUND")
}

func TestDeleteColumn_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	bID := createBoard(t, owner.Token, wsID)
	colID := createColumn(t, owner.Token, bID)

	resp := doRequest(t, http.MethodDelete, "/columns/"+colID, nil, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}
