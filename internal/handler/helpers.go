package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"getkanbam.app/api/internal/apierr"
	"getkanbam.app/api/internal/convert"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

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
	var apiErr apierr.APIError
	if errors.As(err, &apiErr) {
		apiErr.RespondTo(w)
		return
	}
	apierr.UnexpectedServerError().RespondTo(w)
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
