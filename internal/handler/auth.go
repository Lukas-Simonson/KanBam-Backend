package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"getkanbam.app/api/internal/apierr"
	"getkanbam.app/api/internal/middleware"
	"getkanbam.app/api/internal/model"
	"getkanbam.app/api/internal/services"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(service *services.AuthService) AuthHandler {
	return AuthHandler{
		auth: service,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var registration model.Registration
	err := json.NewDecoder(r.Body).Decode(&registration)

	if err != nil {
		apierr.MalformedRequestBody().RespondTo(w)
		return
	}

	token, err := h.auth.Register(r.Context(), registration)
	if err != nil {
		if apiError, ok := errors.AsType[apierr.APIError](err); ok {
			apiError.RespondTo(w)
			return
		}

		apierr.UnexpectedServerError().RespondTo(w)
		return
	}

	buffer, err := json.Marshal(token)
	if err != nil {
		apierr.UnexpectedServerError().RespondTo(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(buffer)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	email, password, ok := r.BasicAuth()
	if !ok {
		apierr.InvalidAuth().RespondTo(w)
		return
	}

	token, err := h.auth.Login(r.Context(), email, password)
	if err != nil {
		if apiError, ok := errors.AsType[apierr.APIError](err); ok {
			apiError.RespondTo(w)
			return
		}
		apierr.UnexpectedServerError().RespondTo(w)
		return
	}

	buf, err := json.Marshal(token)
	if err != nil {
		apierr.UnexpectedServerError().RespondTo(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(buf)
}

func (h *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apierr.InvalidAuth().RespondTo(w)
		return
	}

	var body model.UpdatePassword
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierr.MalformedRequestBody().RespondTo(w)
		return
	}

	err := h.auth.UpdatePassword(r.Context(), userID, body.CurrentPassword, body.NewPassword)
	if err != nil {
		if apiError, ok := errors.AsType[apierr.APIError](err); ok {
			apiError.RespondTo(w)
			return
		}
		apierr.UnexpectedServerError().RespondTo(w)
		return
	}

	w.WriteHeader(http.StatusOK)
}
