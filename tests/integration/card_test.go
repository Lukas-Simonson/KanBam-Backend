package integration

import (
	"net/http"
	"testing"
)

// --- Create Card (board-scoped) ---

func TestCreateCard_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/cards/create", map[string]any{
		"title":    "My Card",
		"position": 1.0,
	}, user.Token)
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "id")
	mustHaveField(t, out, "reference")
}

func TestCreateCard_WithColumn(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	colID := createColumn(t, user.Token, bID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/cards/create", map[string]any{
		"title":    "Card in Column",
		"position": 1.0,
		"columnID": colID,
	}, user.Token)
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["columnID"] != colID {
		t.Errorf("columnID mismatch: got %v, want %v", out["columnID"], colID)
	}
}

func TestCreateCard_MissingTitle(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/cards/create", map[string]any{
		"position": 1.0,
	}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestCreateCard_ViewerRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	bID := createBoard(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/cards/create", map[string]any{
		"title":    "Viewer Card",
		"position": 1.0,
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

func TestCreateCard_BoardNotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPost, "/boards/"+nonexistentUUID+"/cards/create", map[string]any{
		"title":    "Ghost Card",
		"position": 1.0,
	}, user.Token)
	// Board not found returns 404 (from the board lookup in the handler)
	assertStatus(t, resp, http.StatusNotFound)
	drainBody(resp)
}

// --- List Cards (board-scoped) ---

func TestListCards_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	createCard(t, user.Token, bID)
	createCard(t, user.Token, bID)

	resp := doRequest(t, http.MethodGet, "/boards/"+bID+"/cards", nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out struct {
		Cards []map[string]any `json:"cards"`
	}
	decodeBody(t, resp, &out)
	if len(out.Cards) < 2 {
		t.Errorf("expected at least 2 cards, got %d", len(out.Cards))
	}
}

func TestListCards_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)

	resp := doRequest(t, http.MethodGet, "/boards/"+bID+"/cards", nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Get Card ---

func TestGetCard_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)

	resp := doRequest(t, http.MethodGet, "/cards/"+cardID, nil, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["id"] != cardID {
		t.Errorf("id mismatch: got %v, want %v", out["id"], cardID)
	}
}

func TestGetCard_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodGet, "/cards/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "CARD_NOT_FOUND")
}

func TestGetCard_NotMember(t *testing.T) {
	owner := registerUser(t)
	outsider := registerUser(t)
	wsID := createWorkspace(t, owner.Token)
	bID := createBoard(t, owner.Token, wsID)
	cardID := createCard(t, owner.Token, bID)

	resp := doRequest(t, http.MethodGet, "/cards/"+cardID, nil, outsider.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Update Card ---

func TestUpdateCard_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)

	resp := doRequest(t, http.MethodPatch, "/cards/"+cardID, map[string]any{
		"title": "Updated Card",
	}, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["title"] != "Updated Card" {
		t.Errorf("title mismatch: got %v", out["title"])
	}
}

func TestUpdateCard_Archive(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)

	archived := true
	resp := doRequest(t, http.MethodPatch, "/cards/"+cardID, map[string]any{
		"isArchived": archived,
	}, user.Token)
	assertStatus(t, resp, http.StatusOK)

	var out map[string]any
	decodeBody(t, resp, &out)
	if out["isArchived"] != true {
		t.Errorf("expected isArchived=true, got %v", out["isArchived"])
	}
}

func TestUpdateCard_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodPatch, "/cards/"+nonexistentUUID, map[string]any{
		"title": "Ghost",
	}, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "CARD_NOT_FOUND")
}

func TestUpdateCard_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	bID := createBoard(t, owner.Token, wsID)
	cardID := createCard(t, owner.Token, bID)

	resp := doRequest(t, http.MethodPatch, "/cards/"+cardID, map[string]any{
		"title": "Viewer Update",
	}, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Delete Card ---

func TestDeleteCard_Success(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)
	cardID := createCard(t, user.Token, bID)

	resp := doRequest(t, http.MethodDelete, "/cards/"+cardID, nil, user.Token)
	assertStatus(t, resp, http.StatusNoContent)
	drainBody(resp)

	getResp := doRequest(t, http.MethodGet, "/cards/"+cardID, nil, user.Token)
	assertStatus(t, getResp, http.StatusNotFound)
	drainBody(getResp)
}

func TestDeleteCard_NotFound(t *testing.T) {
	user := registerUser(t)
	resp := doRequest(t, http.MethodDelete, "/cards/"+nonexistentUUID, nil, user.Token)
	assertStatus(t, resp, http.StatusNotFound)
	assertErrorCode(t, resp, "CARD_NOT_FOUND")
}

func TestDeleteCard_InsufficientRole(t *testing.T) {
	owner, viewer, wsID := setupWorkspaceWithMember(t, "viewer")
	bID := createBoard(t, owner.Token, wsID)
	cardID := createCard(t, owner.Token, bID)

	resp := doRequest(t, http.MethodDelete, "/cards/"+cardID, nil, viewer.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	drainBody(resp)
}

// --- Sequence / Reference ---

func TestCreateCard_SequenceIncrement(t *testing.T) {
	user := registerUser(t)
	wsID := createWorkspace(t, user.Token)
	bID := createBoard(t, user.Token, wsID)

	var refs []string
	for i := 0; i < 3; i++ {
		resp := doRequest(t, http.MethodPost, "/boards/"+bID+"/cards/create", map[string]any{
			"title":    "Card",
			"position": float64(i + 1),
		}, user.Token)
		assertStatus(t, resp, http.StatusCreated)
		var out map[string]any
		decodeBody(t, resp, &out)
		refs = append(refs, out["reference"].(string))
	}

	// All references should be distinct
	seen := map[string]bool{}
	for _, r := range refs {
		if seen[r] {
			t.Errorf("duplicate reference: %s", r)
		}
		seen[r] = true
	}
}
