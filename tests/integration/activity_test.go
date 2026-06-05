package integration

import (
	"net/http"
	"testing"
)

// --- Workspace Activity ---

func TestGetWorkspaceActivity_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/activity", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Activities []any `json:"activities"`
	}
	decodeBody(t, resp, &out)
	// Activities may be empty; the endpoint should still return 200 with a list
	if out.Activities == nil {
		t.Error("expected activities array (even if empty), got nil")
	}
}

func TestGetWorkspaceActivity_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/activity", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestGetWorkspaceActivity_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/workspaces/"+nonexistentUUID+"/activity", nil, user.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestGetWorkspaceActivity_NoAuth(t *testing.T) {
	owner := registerUser(t)
	wsID := createWorkspace(t, owner.Token)

	resp := doRequest(t, http.MethodGet, "/workspaces/"+wsID+"/activity", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Board Activity ---

func TestGetBoardActivity_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/boards/"+bID+"/activity", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Activities []any `json:"activities"`
	}
	decodeBody(t, resp, &out)
	if out.Activities == nil {
		t.Error("expected activities array (even if empty), got nil")
	}
}

func TestGetBoardActivity_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/boards/"+nonexistentUUID+"/activity", nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "BOARD_NOT_FOUND")
}

func TestGetBoardActivity_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/boards/"+bID+"/activity", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Card Activity ---

func TestGetCardActivity_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)

	resp := doRequest(t, http.MethodGet, "/cards/"+cardID+"/activity", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Activities []any `json:"activities"`
	}
	decodeBody(t, resp, &out)
	if out.Activities == nil {
		t.Error("expected activities array (even if empty), got nil")
	}
}

func TestGetCardActivity_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/cards/"+nonexistentUUID+"/activity", nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "CARD_NOT_FOUND")
}

func TestGetCardActivity_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)
	cardID := createCard(t, owner.Token, bID)

	resp := doRequest(t, http.MethodGet, "/cards/"+cardID+"/activity", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Direct Activity Access ---

func TestGetActivity_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/activity/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "ACTIVITY_NOT_FOUND")
}

func TestGetActivity_NoAuth(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/activity/"+nonexistentUUID, nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}
