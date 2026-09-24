package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	authpkg "armss-poster-studio/internal/auth"
	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/middleware"
	"armss-poster-studio/internal/http/respond"
)

type AuthHandler struct {
	Manager *authpkg.Manager
}

func NewAuthHandler(manager *authpkg.Manager) *AuthHandler {
	return &AuthHandler{Manager: manager}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Request body must be valid JSON with email and password.")
		return
	}

	cookie, user, err := h.Manager.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, authpkg.ErrInvalidCredentials) {
		respond.Error(w, http.StatusUnauthorized, "Email or password is incorrect.")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not log in right now. Please try again.")
		return
	}

	http.SetCookie(w, cookie)
	respond.JSON(w, http.StatusOK, dto.UserResponse{ID: user.ID, Email: user.Email, Role: user.Role})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := ""
	if cookie, err := r.Cookie(h.Manager.CookieName()); err == nil {
		token = cookie.Value
	}
	http.SetCookie(w, h.Manager.Logout(r.Context(), token))
	respond.JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

// Me returns the currently authenticated user, so the SPA can bootstrap its auth
// state on load. Not in the spec's literal endpoint list, but required for a
// session-cookie SPA to know who's logged in without guessing.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "You need to be logged in for this.")
		return
	}
	respond.JSON(w, http.StatusOK, dto.UserResponse{ID: user.ID, Email: user.Email, Role: user.Role})
}
