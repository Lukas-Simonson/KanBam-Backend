package handler

import (
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
	registration, ok := decodeAndValidate[model.Registration](w, r)
	if !ok {
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

	respondJSON(w, http.StatusCreated, token)
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

	respondJSON(w, http.StatusCreated, token)
}

func (h *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apierr.InvalidAuth().RespondTo(w)
		return
	}

	body, ok := decodeAndValidate[model.UpdatePassword](w, r)
	if !ok {
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
