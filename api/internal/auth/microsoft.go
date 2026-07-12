package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MicrosoftAuthService handles Microsoft OAuth authentication
type MicrosoftAuthService struct {
	db           *pgxpool.Pool
	logger       *slog.Logger
	clientID     string
	clientSecret string
	redirectURI  string
	// For demo, store auth codes in memory
	authCodeStore map[string]authCodeData
}

type authCodeData struct {
	userEmail string
	timestamp int64
}

// NewMicrosoftAuthService creates a new Microsoft auth service
func NewMicrosoftAuthService(db *pgxpool.Pool, logger *slog.Logger) *MicrosoftAuthService {
	return &MicrosoftAuthService{
		db:            db,
		logger:        logger,
		authCodeStore: make(map[string]authCodeData),
		clientID:      "", // Would load from env
		clientSecret:  "", // Would load from env
		redirectURI:   "http://localhost:8080/api/v1/auth/microsoft/callback",
	}
}

// StartOAuth returns the Microsoft login URL
func (mas *MicrosoftAuthService) StartOAuth(ctx context.Context) (string, error) {
	// Generate state for CSRF protection
	state := make([]byte, 16)
	if _, err := rand.Read(state); err != nil {
		return "", err
	}
	stateStr := hex.EncodeToString(state)

	// In production, would construct real Microsoft OAuth URL
	// For demo, just return a mock flow
	mas.logger.Info("OAuth flow initiated", "state", stateStr)

	// Return the authorization URL
	params := url.Values{
		"client_id":     {mas.clientID},
		"response_type": {"code"},
		"redirect_uri":  {mas.redirectURI},
		"scope":         {"openid profile email"},
		"state":         {stateStr},
	}

	// Mock URL for demo
	return fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/authorize?%s", params.Encode()), nil
}

// HandleCallback processes the OAuth callback and returns a user ID
func (mas *MicrosoftAuthService) HandleCallback(ctx context.Context, code, state string) (string, error) {
	if code == "" {
		return "", errors.New("missing authorization code")
	}

	// In production, exchange code for token via Microsoft API
	// For demo, parse the mock code to extract email
	userEmail := parseMockAuthCode(code)
	if userEmail == "" {
		return "", errors.New("invalid authorization code")
	}

	// Find or create user by email
	var userID string
	err := mas.db.QueryRow(ctx, `
		SELECT id::text FROM users WHERE email = $1
	`, userEmail).Scan(&userID)

	if err == nil {
		// User exists
		return userID, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	// Create new user
	email := userEmail
	displayName := parseEmailName(userEmail)

	err = mas.db.QueryRow(ctx, `
		INSERT INTO users (email, display_name, status)
		VALUES ($1, $2, 'active')
		RETURNING id::text
	`, email, displayName).Scan(&userID)

	return userID, err
}

// Helper functions

func parseMockAuthCode(code string) string {
	// Demo: code format is "mock_<email_base64>"
	// In real implementation, exchange with Microsoft API
	if len(code) > 5 && code[:5] == "mock_" {
		// For demo, just use a fake email
		return "user@example.com"
	}
	return ""
}

func parseEmailName(email string) string {
	for i, c := range email {
		if c == '@' {
			return email[:i]
		}
	}
	return email
}

// HTTP Handlers

type microsoftCallbackQuery struct {
	Code  string
	State string
}

func (mas *MicrosoftAuthService) StartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oauthURL, err := mas.StartOAuth(r.Context())
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "OAUTH_ERROR", err.Error())
			return
		}

		// Redirect to Microsoft login
		http.Redirect(w, r, oauthURL, http.StatusFound)
	}
}

func (mas *MicrosoftAuthService) CallbackHandler(authSvc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" {
			error_ := r.URL.Query().Get("error")
			writeError(w, http.StatusUnauthorized, "OAUTH_DENIED", error_)
			return
		}

		userID, err := mas.HandleCallback(r.Context(), code, state)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "OAUTH_ERROR", err.Error())
			return
		}

		// Issue session
		token, expires, err := authSvc.IssueSession(r.Context(), userID, false, r.UserAgent(), getClientIP(r))
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "SESSION_ERROR", "Failed to create session.")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    token,
			Path:     "/",
			Expires:  expires,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   r.TLS != nil,
		})

		// Redirect to app
		http.Redirect(w, r, "http://localhost:5173?session="+token, http.StatusFound)
	}
}
