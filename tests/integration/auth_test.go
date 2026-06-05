package integration

import (
	"net/http"
	"testing"
)

func TestRegister_Success(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    uniqueEmail(),
		"name":     "TestUser",
		"password": "Secret1!",
	}, "")
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "token")
}

func TestRegister_DuplicateEmail(t *testing.T) {
	email := uniqueEmail()
	doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    email,
		"name":     "First",
		"password": "Secret1!",
	}, "").Body.Close()

	resp := doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    email,
		"name":     "Second",
		"password": "Secret1!",
	}, "")
	assertStatus(t, resp, http.StatusConflict)
	assertErrorCode(t, resp, "EMAIL_TAKEN")
}

func TestRegister_MissingEmail(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"name":     "NoEmail",
		"password": "Secret1!",
	}, "")
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestRegister_InvalidEmail(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    "not-an-email",
		"name":     "BadEmail",
		"password": "Secret1!",
	}, "")
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestRegister_WeakPassword(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    uniqueEmail(),
		"name":     "WeakPass",
		"password": "password",
	}, "")
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestRegister_NameTooShort(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    uniqueEmail(),
		"name":     "A",
		"password": "Secret1!",
	}, "")
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}

func TestRegister_MalformedJSON(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/auth/register", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "MALFORMED_REQUEST_BODY")
}

func TestLogin_Success(t *testing.T) {
	email := uniqueEmail()
	doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    email,
		"name":     "LoginUser",
		"password": "Secret1!",
	}, "").Body.Close()

	resp := doBasicAuth(t, email, "Secret1!")
	assertStatus(t, resp, http.StatusCreated)

	var out map[string]any
	decodeBody(t, resp, &out)
	mustHaveField(t, out, "token")
}

func TestLogin_WrongPassword(t *testing.T) {
	email := uniqueEmail()
	doRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"email":    email,
		"name":     "WrongPass",
		"password": "Secret1!",
	}, "").Body.Close()

	resp := doBasicAuth(t, email, "WrongPassword1!")
	assertStatus(t, resp, http.StatusUnauthorized)
	assertErrorCode(t, resp, "INVALID_CREDENTIALS")
}

func TestLogin_UnknownEmail(t *testing.T) {
	resp := doBasicAuth(t, "nobody@example.com", "Secret1!")
	assertStatus(t, resp, http.StatusUnauthorized)
	assertErrorCode(t, resp, "INVALID_CREDENTIALS")
}

func TestLogin_NoBasicAuth(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/auth/login", nil, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	assertErrorCode(t, resp, "INVALID_AUTH")
}

func TestUpdatePassword_Success(t *testing.T) {
	user := registerUser(t)

	resp := doRequest(t, http.MethodPost, "/auth/updatePassword", map[string]any{
		"currentPassword": "Secret1!",
		"newPassword":     "NewSecret2@",
	}, user.Token)
	assertStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// old password no longer works
	oldResp := doBasicAuth(t, user.Email, "Secret1!")
	assertStatus(t, oldResp, http.StatusUnauthorized)
	oldResp.Body.Close()

	// new password works
	newResp := doBasicAuth(t, user.Email, "NewSecret2@")
	assertStatus(t, newResp, http.StatusCreated)
	newResp.Body.Close()
}

func TestUpdatePassword_WrongCurrentPassword(t *testing.T) {
	user := registerUser(t)

	resp := doRequest(t, http.MethodPost, "/auth/updatePassword", map[string]any{
		"currentPassword": "Wrong1!",
		"newPassword":     "NewSecret2@",
	}, user.Token)
	assertStatus(t, resp, http.StatusUnauthorized)
	assertErrorCode(t, resp, "INVALID_PASSWORD")
}

func TestUpdatePassword_NoAuth(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/auth/updatePassword", map[string]any{
		"currentPassword": "Secret1!",
		"newPassword":     "NewSecret2@",
	}, "")
	assertStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestUpdatePassword_WeakNewPassword(t *testing.T) {
	user := registerUser(t)

	resp := doRequest(t, http.MethodPost, "/auth/updatePassword", map[string]any{
		"currentPassword": "Secret1!",
		"newPassword":     "weak",
	}, user.Token)
	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorCode(t, resp, "INVALID_REQUEST")
}
