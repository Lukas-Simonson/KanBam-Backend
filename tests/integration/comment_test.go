package integration

import (
	"net/http"
	"testing"
)

// --- Create Comment (card-scoped) ---

func TestCreateComment_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)

	resp := doRequest(t, http.MethodPost, "/cards/"+cardID+"/comments/create", map[string]any{
		"body": "This is a comment",
	}, user.Token)
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "id")
	if out["body"] != "This is a comment" {
		t.Errorf("body mismatch: got %v", out["body"])
	}
}

func TestCreateComment_MissingBody(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)

	resp := doRequest(t, http.MethodPost, "/cards/"+cardID+"/comments/create", map[string]any{}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestCreateComment_NoAuth(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)

	resp := doRequest(t, http.MethodPost, "/cards/"+cardID+"/comments/create", map[string]any{
		"body": "Unauthorized",
	}, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- List Comments (card-scoped) ---

func TestListComments_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)
	createComment(t, user.Token, cardID)
	createComment(t, user.Token, cardID)

	resp := doRequest(t, http.MethodGet, "/cards/"+cardID+"/comments", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Comments []map[string]any `json:"comments"`
	}
	decodeBody(t, resp, &out)
	if len(out.Comments) < 2 {
		t.Errorf("expected at least 2 comments, got %d", len(out.Comments))
	}
}

func TestListComments_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)
	cardID := createCard(t, owner.Token, bID)

	resp := doRequest(t, http.MethodGet, "/cards/"+cardID+"/comments", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Get Comment ---

func TestGetComment_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)
	commentID := createComment(t, user.Token, cardID)

	resp := doRequest(t, http.MethodGet, "/comments/"+commentID, nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["id"] != commentID {
		t.Errorf("id mismatch: got %v, want %v", out["id"], commentID)
	}
}

func TestGetComment_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/comments/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "COMMENT_NOT_FOUND")
}

func TestGetComment_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)
	cardID := createCard(t, owner.Token, bID)
	commentID := createComment(t, owner.Token, cardID)

	resp := doRequest(t, http.MethodGet, "/comments/"+commentID, nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Update Comment ---

func TestUpdateComment_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)
	commentID := createComment(t, user.Token, cardID)

	resp := doRequest(t, http.MethodPatch, "/comments/"+commentID, map[string]any{
		"body": "Updated comment body",
	}, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["body"] != "Updated comment body" {
		t.Errorf("body mismatch: got %v", out["body"])
	}
}

func TestUpdateComment_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPatch, "/comments/"+nonexistentUUID, map[string]any{
		"body": "Ghost",
	}, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "COMMENT_NOT_FOUND")
}

func TestUpdateComment_NotAuthor(t *testing.T) {
	owner, contributer, wsID := setupWorkspaceWithMember(t, "contributer")
	bID := createBoard(t, owner.Token, wsID)
	cardID := createCard(t, owner.Token, bID)
	commentID := createComment(t, owner.Token, cardID)

	// Another member tries to update a comment they didn't write
	resp := doRequest(t, http.MethodPatch, "/comments/"+commentID, map[string]any{
		"body": "Tampered",
	}, contributer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Delete Comment ---

func TestDeleteComment_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)
	commentID := createComment(t, user.Token, cardID)

	resp := doRequest(t, http.MethodDelete, "/comments/"+commentID, nil, user.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)

	getResp := doRequest(t, http.MethodGet, "/comments/"+commentID, nil, user.Token)
	assertStatus(t, getResp, http.StatusNotFound)
	drainBody(getResp)
}

func TestDeleteComment_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodDelete, "/comments/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "COMMENT_NOT_FOUND")
}

func TestDeleteComment_NotAuthor(t *testing.T) {
	owner, contributer, wsID := setupWorkspaceWithMember(t, "contributer")
	bID := createBoard(t, owner.Token, wsID)
	cardID := createCard(t, owner.Token, bID)
	commentID := createComment(t, owner.Token, cardID)

	resp := doRequest(t, http.MethodDelete, "/comments/"+commentID, nil, contributer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}
