package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
)

// counter provides unique suffixes for test data to prevent collisions between tests.
var counter atomic.Int64

func uniqueSuffix() string {
	return fmt.Sprintf("%d", counter.Add(1))
}

func uniqueEmail() string {
	return fmt.Sprintf("user%s@example.com", uniqueSuffix())
}

// doRequest sends an HTTP request and returns the response.
func doRequest(t *testing.T, method, path string, body any, token string) *http.Response {
	t.Helper()
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, srv.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func doBasicAuth(t *testing.T, email, password string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/auth/login", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.SetBasicAuth(email, password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

// decodeBody decodes the JSON response body into dst.
func decodeBody(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

// assertStatus checks the response status code.
func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		t.Errorf("status %d, want %d; body: %v", resp.StatusCode, want, body)
	}
}

// assertErrorCode decodes the response body and checks the error code field.
func assertErrorCode(t *testing.T, resp *http.Response, wantCode string) {
	t.Helper()
	var body struct {
		Code string `json:"code"`
	}
	decodeBody(t, resp, &body)
	if body.Code != wantCode {
		t.Errorf("error code %q, want %q", body.Code, wantCode)
	}
}

// --- Seed helpers ---

type userCredentials struct {
	ID    string
	Email string
	Token string
}

// registerUser registers a new user and returns their credentials.
func registerUser(t *testing.T) userCredentials {
	t.Helper()
	email := uniqueEmail()
	n := uniqueSuffix()
	name := "User" + n
	resp := doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    email,
		"name":     name,
		"password": "Secret1!",
	}, "")
	assertStatus(t, resp, http.StatusCreated)
	var out struct {
		Token string `json:"token"`
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	decodeBody(t, resp, &out)
	if out.Token == "" {
		t.Fatal("expected token in register response")
	}
	return userCredentials{ID: out.User.ID, Email: email, Token: out.Token}
}

// createWorkspace creates a workspace and returns its ID.
func createWorkspace(t *testing.T, token string) string {
	t.Helper()
	n := uniqueSuffix()
	resp := doRequest(t, http.MethodPost, "/workspaces/create", map[string]any{
		"title": "Workspace " + n,
	}, token)
	assertStatus(t, resp, http.StatusCreated)
	var out struct {
		ID string `json:"id"`
	}
	decodeBody(t, resp, &out)
	if out.ID == "" {
		t.Fatal("expected workspace ID in response")
	}
	return out.ID
}

// createBoard creates a board inside a workspace and returns its ID.
func createBoard(t *testing.T, token, workspaceID string) string {
	t.Helper()
	n := uniqueSuffix()
	prefix := "B" + n
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	resp := doRequest(t, http.MethodPost, "/workspaces/"+workspaceID+"/boards/create", map[string]any{
		"title":  "Board " + n,
		"prefix": prefix,
	}, token)
	assertStatus(t, resp, http.StatusCreated)
	var out struct {
		ID string `json:"id"`
	}
	decodeBody(t, resp, &out)
	if out.ID == "" {
		t.Fatal("expected board ID in response")
	}
	return out.ID
}

// createColumn creates a column inside a board and returns its ID.
func createColumn(t *testing.T, token, boardID string) string {
	t.Helper()
	n := uniqueSuffix()
	resp := doRequest(t, http.MethodPost, "/boards/"+boardID+"/columns/create", map[string]any{
		"title":    "Column " + n,
		"position": 0,
	}, token)
	assertStatus(t, resp, http.StatusCreated)
	var out struct {
		ID string `json:"id"`
	}
	decodeBody(t, resp, &out)
	if out.ID == "" {
		t.Fatal("expected column ID in response")
	}
	return out.ID
}

// createCard creates a card inside a board and returns its ID.
func createCard(t *testing.T, token, boardID string) string {
	t.Helper()
	n := uniqueSuffix()
	resp := doRequest(t, http.MethodPost, "/boards/"+boardID+"/cards/create", map[string]any{
		"title":    "Card " + n,
		"position": 1.0,
	}, token)
	assertStatus(t, resp, http.StatusCreated)
	var out struct {
		ID string `json:"id"`
	}
	decodeBody(t, resp, &out)
	if out.ID == "" {
		t.Fatal("expected card ID in response")
	}
	return out.ID
}

// createComment creates a comment on a card and returns its ID.
func createComment(t *testing.T, token, cardID string) string {
	t.Helper()
	n := uniqueSuffix()
	resp := doRequest(t, http.MethodPost, "/cards/"+cardID+"/comments/create", map[string]any{
		"body": "Comment " + n,
	}, token)
	assertStatus(t, resp, http.StatusCreated)
	var out struct {
		ID string `json:"id"`
	}
	decodeBody(t, resp, &out)
	if out.ID == "" {
		t.Fatal("expected comment ID in response")
	}
	return out.ID
}

// createTag creates a tag in a workspace and returns its ID.
func createTag(t *testing.T, token, workspaceID string) string {
	t.Helper()
	n := uniqueSuffix()
	resp := doRequest(t, http.MethodPost, "/workspaces/"+workspaceID+"/tags/create", map[string]any{
		"name":  "Tag " + n,
		"color": "#FF0000",
	}, token)
	assertStatus(t, resp, http.StatusCreated)
	var out struct {
		ID string `json:"id"`
	}
	decodeBody(t, resp, &out)
	if out.ID == "" {
		t.Fatal("expected tag ID in response")
	}
	return out.ID
}

// nonexistentUUID is a valid UUID that won't exist in the database.
const nonexistentUUID = "00000000-0000-4000-8000-000000000001"

// addMember adds userID to workspaceID as the given role using the caller's token.
func addMember(t *testing.T, callerToken, workspaceID, userID, role string) {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/workspaces/"+workspaceID+"/members", map[string]any{
		"userID": userID,
		"role":   role,
	}, callerToken)
	if resp.StatusCode != http.StatusNoContent {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		t.Fatalf("addMember: status %d; body: %v", resp.StatusCode, body)
	}
	resp.Body.Close()
}

// setupWorkspaceWithMember creates a workspace owner and a second member with the given role.
func setupWorkspaceWithMember(t *testing.T, memberRole string) (owner userCredentials, member userCredentials, workspaceID string) {
	t.Helper()
	owner = registerUser(t)
	member = registerUser(t)
	workspaceID = createWorkspace(t, owner.Token)
	addMember(t, owner.Token, workspaceID, member.ID, memberRole)
	return owner, member, workspaceID
}

// drainBody closes the response body without reading it.
func drainBody(resp *http.Response) {
	resp.Body.Close()
}

// mustHaveField checks that a decoded map has a non-empty field.
func mustHaveField(t *testing.T, m map[string]any, field string) {
	t.Helper()
	v, ok := m[field]
	if !ok || v == nil || v == "" {
		t.Errorf("expected non-empty field %q in response; got %v", field, m)
	}
}
