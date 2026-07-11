package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Service) Me(w http.ResponseWriter, r *http.Request) {
	user, err := s.CurrentUser(r.Context(), cookieValue(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication is required.")
		return
	}
	writeData(w, http.StatusOK, map[string]any{"id": user.ID, "email": user.Email, "phone": user.Phone, "displayName": user.DisplayName, "isAdmin": user.IsAdmin})
}

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) {
	_ = s.Revoke(r.Context(), cookieValue(r))
	http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: r.TLS != nil})
	w.WriteHeader(http.StatusNoContent)
}

// Phone and Microsoft provider exchanges are intentionally backend-only. These
// endpoints make the active authentication mode explicit until their provider
// credentials are configured in Phase 2.
func AuthModeGuard(mode, expected string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if mode != expected {
			writeError(w, http.StatusNotFound, "AUTH_MODE_DISABLED", "This authentication method is disabled.")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "AUTH_PROVIDER_NOT_CONFIGURED", "Authentication provider setup is incomplete.")
	}
}
func cookieValue(r *http.Request) string {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}
func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data, "requestId": ""})
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": code, "message": message, "details": map[string]any{}}, "requestId": ""})
}
