package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"getkanbam.app/api/internal/apierr"
	"getkanbam.app/api/internal/convert"
	"getkanbam.app/api/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("password_complexity", validatePasswordComplexity)
}

func validatePasswordComplexity(fl validator.FieldLevel) bool {
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range fl.Field().String() {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

// decodeAndValidate decodes a JSON body into T and runs struct validation.
// On failure it writes the appropriate error response and returns false.
func decodeAndValidate[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		apierr.MalformedRequestBody().RespondTo(w)
		return v, false
	}
	if err := validate.Struct(v); err != nil {
		apierr.InvalidRequest(validationReason(err)).RespondTo(w)
		return v, false
	}
	return v, true
}

func validationReason(err error) string {
	if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
		parts := make([]string, 0, len(ve))
		for _, e := range ve {
			parts = append(parts, fmt.Sprintf("%s failed %s validation", e.Field(), e.Tag()))
		}
		return strings.Join(parts, "; ")
	}
	return err.Error()
}

// respondJSON marshals v and writes it with the given status code.
func respondJSON(w http.ResponseWriter, status int, v any) {
	buf, err := json.Marshal(v)
	if err != nil {
		apierr.UnexpectedServerError().RespondTo(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf)
}

// handleError maps a service error to an API response.
// APIErrors are written directly; everything else becomes a 500.
func handleError(w http.ResponseWriter, err error) {
	if apiErr, ok := errors.AsType[apierr.APIError](err); ok {
		apiErr.RespondTo(w)
		return
	}
	apierr.UnexpectedServerError().RespondTo(w)
}

// mustCallerID extracts the authenticated user's ID from the request context.
// It writes a 401 and returns false if the ID is missing (should not happen under JWTBearerMiddleware).
func mustCallerID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, ok := middleware.GetUserID(r.Context())
	if !ok {
		apierr.InvalidAuth().RespondTo(w)
	}
	return id, ok
}

// urlUUID extracts a named URL parameter and parses it as a UUID.
// Returns false and writes a 400 if the parameter is missing or invalid.
func urlUUID(w http.ResponseWriter, r *http.Request, param string) (pgtype.UUID, bool) {
	raw := chi.URLParam(r, param)
	id, err := convert.ParseUUID(raw)
	if err != nil {
		apierr.MalformedRequestBody().RespondTo(w)
		return pgtype.UUID{}, false
	}
	return id, true
}
