package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"ticket-system/internal/models"
	"ticket-system/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userResp, err := h.authService.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail), errors.Is(err, service.ErrPasswordTooShort):
			RespondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrUserAlreadyExists):
			RespondError(w, http.StatusConflict, err.Error())
		default:
			RespondError(w, http.StatusInternalServerError, "failed to register user")
		}
		return
	}

	RespondJSON(w, http.StatusCreated, userResp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	authResp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			RespondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		RespondError(w, http.StatusInternalServerError, "failed to authenticate user")
		return
	}

	RespondJSON(w, http.StatusOK, authResp)
}
