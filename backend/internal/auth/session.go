package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"armss-poster-studio/internal/store/mysql"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type Manager struct {
	users      *mysql.UsersRepo
	sessions   *mysql.SessionsRepo
	cookieName string
	ttl        time.Duration
	secure     bool
}

func NewManager(users *mysql.UsersRepo, sessions *mysql.SessionsRepo, cookieName string, ttl time.Duration, secure bool) *Manager {
	return &Manager{users: users, sessions: sessions, cookieName: cookieName, ttl: ttl, secure: secure}
}

// Login verifies credentials and creates a new server-side session, returning the
// session cookie to set and the authenticated user.
func (m *Manager) Login(ctx context.Context, email, password string) (*http.Cookie, *mysql.User, error) {
	user, err := m.users.GetByEmail(ctx, email)
	if errors.Is(err, mysql.ErrNotFound) {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, fmt.Errorf("lookup user: %w", err)
	}
	if !ComparePassword(user.PasswordHash, password) {
		return nil, nil, ErrInvalidCredentials
	}

	token, err := randomToken()
	if err != nil {
		return nil, nil, fmt.Errorf("generate session token: %w", err)
	}
	expiresAt := time.Now().Add(m.ttl)
	if err := m.sessions.Create(ctx, token, user.ID, expiresAt); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	return &http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	}, user, nil
}

// Logout deletes the server-side session and returns a cookie that clears it client-side.
func (m *Manager) Logout(ctx context.Context, token string) *http.Cookie {
	if token != "" {
		_ = m.sessions.Delete(ctx, token)
	}
	return &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
}

// UserFromToken resolves a session cookie value to its owning user.
func (m *Manager) UserFromToken(ctx context.Context, token string) (*mysql.User, error) {
	sess, err := m.sessions.Get(ctx, token)
	if errors.Is(err, mysql.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("lookup session: %w", err)
	}
	return m.users.GetByID(ctx, sess.UserID)
}

func (m *Manager) CookieName() string { return m.cookieName }

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
