package httpapi

import (
	"encoding/json"
	"github.com/arryaanjain/AI_DAY/internal/assets"
	"github.com/arryaanjain/AI_DAY/internal/auth"
	"github.com/arryaanjain/AI_DAY/internal/credits"
	"github.com/arryaanjain/AI_DAY/internal/generation"
	"net/http"
)

type generationRequest struct {
	SourceAssetID   string `json:"sourceAssetId"`
	IdempotencyKey  string `json:"idempotencyKey"`
	BiggestHigh     string `json:"biggestHigh"`
	BiggestLow      string `json:"biggestLow"`
	ProtagonistName string `json:"protagonistName"`
	Language        string `json:"language"`
	Tone            string `json:"tone"`
}

func authenticatedUser(w http.ResponseWriter, r *http.Request, s *auth.Service) (auth.User, bool) {
	cookie, cookieErr := r.Cookie(auth.SessionCookieName)
	if cookieErr != nil {
		errorResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication is required.")
		return auth.User{}, false
	}
	user, err := s.CurrentUser(r.Context(), cookie.Value)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication is required.")
		return auth.User{}, false
	}
	return user, true
}
func creditBalanceHandler(a *auth.Service, c *credits.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, a)
		if !ok {
			return
		}
		balance, err := c.Balance(r.Context(), user.ID)
		if err != nil {
			errorResponse(w, http.StatusServiceUnavailable, "INTERNAL_ERROR", "Credit balance is temporarily unavailable.")
			return
		}
		respond(w, http.StatusOK, map[string]int{"available": balance})
	}
}
func createGenerationHandler(module string, a *auth.Service, c *credits.Service, assetsService *assets.Service, g *generation.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, a)
		if !ok {
			return
		}
		var request generationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.SourceAssetID == "" || request.IdempotencyKey == "" {
			errorResponse(w, http.StatusBadRequest, "INVALID_ASSET", "A source asset and idempotency key are required.")
			return
		}
		if err := assetsService.ConfirmOwnership(r.Context(), user.ID, request.SourceAssetID); err != nil {
			errorResponse(w, http.StatusForbidden, "ASSET_NOT_OWNED", "That source asset is unavailable.")
			return
		}
		if module == "comic" && (request.BiggestHigh == "" || request.BiggestLow == "") {
			errorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Both life events are required for a comic.")
			return
		}
		input, _ := json.Marshal(request)
		jobID, err := g.CreateWithCredit(r.Context(), user.ID, module, request.SourceAssetID, request.IdempotencyKey, input)
		if err != nil {
			if err == generation.ErrInsufficientCredits {
				errorResponse(w, http.StatusPaymentRequired, "INSUFFICIENT_CREDITS", "A generation credit is required.")
				return
			}
			errorResponse(w, http.StatusServiceUnavailable, "INTERNAL_ERROR", "Unable to create generation job.")
			return
		}
		respond(w, http.StatusAccepted, map[string]string{"jobId": jobID, "status": "queued"})
	}
}
func errorResponse(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": code, "message": message, "details": map[string]any{}}, "requestId": ""})
}
