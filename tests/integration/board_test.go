package integration

import (
	"net/http"
	"testing"
)

// --- Create Board ---

func TestCreateBoard_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	n := uniqueSuffix()
	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/boards/create", map[string]any{
		"title":  "My Board",
		"prefix": "MB" + n,
	}, user.Token)
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "id")
	mustHaveField(t, out, "prefix")
}

func TestCreateBoard_DuplicatePrefix(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	n := uniqueSuffix()
	prefix := "DP" + n

	doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/boards/create", map[string]any{
		"title":  "First Board",
		"prefix": prefix,
	}, user.Token).Body.Close()

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/boards/create", map[string]any{
		"title":  "Second Board",
		"prefix": prefix,
	}, user.Token)
	assertStatus(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "BOARD_PREFIX_TAKEN")
}

func TestCreateBoard_MissingTitle(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/boards/create", map[string]any{
		"prefix": "NT" + uniqueSuffix(),
	}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestCreateBoard_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/boards/create", map[string]any{
		"title":  "Sneaky Board",
		"prefix": "SN" + uniqueSuffix(),
	}, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestCreateBoard_ViewerRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	_ = owner

	resp := doRequest(t, http.MethodPost, "/workspaces/"+wsID+"/boards/create", map[string]any{
		"title":  "Viewer Board",
		"prefix": "VB" + uniqueSuffix(),
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- List Boards ---

func TestListBoards_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	createBoard(t, user.Token, wsID)
	createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/boards", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Boards []map[string]any `json:"boards"`
	}
	decodeBody(t, resp, &out)
	if len(out.Boards) < 2 {
		t.Errorf("expected at least 2 boards, got %d", len(out.Boards))
	}
}

func TestListBoards_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/boards", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Get Board ---

func TestGetBoard_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/boards/"+bID, nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["id"] != bID {
		t.Errorf("id mismatch: got %v, want %v", out["id"], bID)
	}
}

func TestGetBoard_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/boards/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "BOARD_NOT_FOUND")
}

func TestGetBoard_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/boards/"+bID, nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Update Board ---

func TestUpdateBoard_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodPatch, "/boards/"+bID, map[string]any{
		"title": "Updated Board",
	}, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["title"] != "Updated Board" {
		t.Errorf("title mismatch: got %v", out["title"])
	}
}

func TestUpdateBoard_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPatch, "/boards/"+nonexistentUUID, map[string]any{
		"title": "Ghost",
	}, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "BOARD_NOT_FOUND")
}

func TestUpdateBoard_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	bID := createBoard(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodPatch, "/boards/"+bID, map[string]any{
		"title": "Viewer Update",
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Delete Board ---

func TestDeleteBoard_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodDelete, "/boards/"+bID, nil, user.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)

	getResp := doRequest(t, http.MethodGet, "/boards/"+bID, nil, user.Token)
	assertStatus(t, getResp, http.StatusNotFound)
	drainBody(getResp)
}

func TestDeleteBoard_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodDelete, "/boards/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "BOARD_NOT_FOUND")
}

func TestDeleteBoard_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	bID := createBoard(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodDelete, "/boards/"+bID, nil, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}
