package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const SessionCookieName = "ai_day_session"

var ErrUnauthenticated = errors.New("unauthenticated")

type User struct {
	ID, Email, Phone, DisplayName string
	IsAdmin                       bool
}
type Service struct {
	db              *pgxpool.Pool
	sessionDuration time.Duration
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db, sessionDuration: 30 * 24 * time.Hour}
}

func (s *Service) CurrentUser(ctx context.Context, rawToken string) (User, error) {
	if s.db == nil || rawToken == "" {
		return User{}, ErrUnauthenticated
	}
	var user User
	err := s.db.QueryRow(ctx, `SELECT u.id::text, COALESCE(u.email::text, ''), COALESCE(u.phone_e164, ''), COALESCE(u.display_name, ''), s.is_admin FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > NOW() AND u.status = 'active'`, hash(rawToken)).Scan(&user.ID, &user.Email, &user.Phone, &user.DisplayName, &user.IsAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUnauthenticated
	}
	return user, err
}

func (s *Service) IssueSession(ctx context.Context, userID string, isAdmin bool, userAgent, ip string) (string, time.Time, error) {
	if s.db == nil {
		return "", time.Time{}, errors.New("database unavailable")
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(bytes)
	expires := time.Now().Add(s.sessionDuration)
	_, err := s.db.Exec(ctx, `INSERT INTO sessions (user_id, token_hash, is_admin, user_agent, ip_address, expires_at) VALUES ($1, $2, $3, $4, NULLIF($5, '')::inet, $6)`, userID, hash(token), isAdmin, userAgent, ip, expires)
	return token, expires, err
}

func (s *Service) Revoke(ctx context.Context, rawToken string) error {
	if s.db == nil || rawToken == "" {
		return nil
	}
	_, err := s.db.Exec(ctx, `UPDATE sessions SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`, hash(rawToken))
	return err
}
func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
