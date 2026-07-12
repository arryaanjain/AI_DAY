package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PhoneAuthService handles phone-based OTP authentication
type PhoneAuthService struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// In production, these would be configured from environment variables
	msg91APIKey string
	templateID  string
	// For demo, we'll use a mock OTP store
	otpStore map[string]storedOTP
}

type storedOTP struct {
	code      string
	expiresAt time.Time
	attempts  int
}

// NewPhoneAuthService creates a new phone auth service
func NewPhoneAuthService(db *pgxpool.Pool, logger *slog.Logger) *PhoneAuthService {
	return &PhoneAuthService{
		db:          db,
		logger:      logger,
		otpStore:    make(map[string]storedOTP),
		msg91APIKey: "", // Would load from env
		templateID:  "", // Would load from env
	}
}

// SendOTP sends an OTP to the phone number
func (pas *PhoneAuthService) SendOTP(ctx context.Context, phone string) error {
	if !isValidPhone(phone) {
		return errors.New("invalid phone number")
	}

	// Generate a 6-digit OTP
	otp := generateOTP()

	// Store OTP in memory (in production, use Redis or DB)
	pas.otpStore[phone] = storedOTP{
		code:      otp,
		expiresAt: time.Now().Add(10 * time.Minute),
		attempts:  0,
	}

	// In production, call MSG91 API to send SMS
	// For demo, just log it
	pas.logger.Info("OTP sent", "phone", maskPhone(phone), "otp", otp)

	return nil
}

// VerifyOTP verifies the OTP and returns user ID
func (pas *PhoneAuthService) VerifyOTP(ctx context.Context, phone, code string) (string, error) {
	stored, ok := pas.otpStore[phone]
	if !ok {
		return "", errors.New("no OTP found for this phone")
	}

	if time.Now().After(stored.expiresAt) {
		delete(pas.otpStore, phone)
		return "", errors.New("OTP expired")
	}

	if stored.attempts >= 3 {
		delete(pas.otpStore, phone)
		return "", errors.New("too many failed attempts")
	}

	if stored.code != code {
		stored.attempts++
		pas.otpStore[phone] = stored
		return "", errors.New("invalid OTP")
	}

	// OTP is valid, find or create user
	delete(pas.otpStore, phone)

	var userID string
	err := pas.db.QueryRow(ctx, `
		SELECT id::text FROM users WHERE phone_e164 = $1
	`, phone).Scan(&userID)

	if err == nil {
		// User exists
		return userID, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	// Create new user
	err = pas.db.QueryRow(ctx, `
		INSERT INTO users (phone_e164, display_name, status)
		VALUES ($1, 'User', 'active')
		RETURNING id::text
	`, phone).Scan(&userID)

	return userID, err
}

// Helper functions

func generateOTP() string {
	b := make([]byte, 3)
	rand.Read(b)
	// Convert to 6-digit OTP (0-999999)
	val := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", val%1000000)
}

func isValidPhone(phone string) bool {
	// Simple validation: starts with +, followed by 1-15 digits
	matched, _ := regexp.MatchString(`^\+\d{10,15}$`, phone)
	return matched
}

func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-2:]
}

// HTTP Handlers

type sendOTPRequest struct {
	Phone string `json:"phone"`
}

type verifyOTPRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func (pas *PhoneAuthService) SendOTPHandler(authSvc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req sendOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Phone number is required.")
			return
		}

		if err := pas.SendOTP(r.Context(), req.Phone); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PHONE", err.Error())
			return
		}

		writeData(w, http.StatusOK, map[string]string{"status": "otp_sent"})
	}
}

func (pas *PhoneAuthService) ResendOTPHandler(authSvc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Same as SendOTP for this implementation
		pas.SendOTPHandler(authSvc)(w, r)
	}
}

func (pas *PhoneAuthService) VerifyOTPHandler(authSvc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req verifyOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" || req.Code == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Phone and OTP code are required.")
			return
		}

		userID, err := pas.VerifyOTP(r.Context(), req.Phone, req.Code)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "INVALID_OTP", err.Error())
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

		writeData(w, http.StatusOK, map[string]any{"userId": userID, "expiresAt": expires})
	}
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fall back to RemoteAddr
	if addr := strings.Split(r.RemoteAddr, ":"); len(addr) > 0 {
		return addr[0]
	}

	return ""
}
