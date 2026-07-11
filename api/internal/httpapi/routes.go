package httpapi

import (
	"encoding/json"
	"github.com/arryaanjain/AI_DAY/internal/assets"
	"github.com/arryaanjain/AI_DAY/internal/auth"
	"github.com/arryaanjain/AI_DAY/internal/config"
	"github.com/arryaanjain/AI_DAY/internal/credits"
	"github.com/arryaanjain/AI_DAY/internal/generation"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"time"
)

type envelope struct {
	Data      any    `json:"data"`
	RequestID string `json:"requestId"`
}

func NewRouter(cfg config.Config, db *pgxpool.Pool, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(requestLogger(logger))
	a := auth.NewService(db)
	c := credits.New(db)
	g := generation.New(db)
	s := assets.New(db, nil, "ai-day", 10*1024*1024, []string{"image/jpeg", "image/png", "image/webp"})
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			respond(w, http.StatusOK, map[string]string{"status": "ok"})
		})
		r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
			if db == nil || db.Ping(r.Context()) != nil {
				respond(w, http.StatusServiceUnavailable, map[string]string{"status": "database_unavailable"})
				return
			}
			respond(w, http.StatusOK, map[string]string{"status": "ready"})
		})
		r.Get("/config/public", func(w http.ResponseWriter, r *http.Request) {
			respond(w, http.StatusOK, map[string]any{"authMode": cfg.AuthMode, "pricePerGenerationPaise": cfg.PricePerGenerationPaise, "currency": cfg.Currency, "enabledModules": []string{"pixart", "comic"}})
		})
		r.Get("/auth/me", a.Me)
		r.Post("/auth/logout", a.Logout)
		r.Post("/auth/phone/send-otp", auth.AuthModeGuard(cfg.AuthMode, "phone"))
		r.Post("/auth/phone/resend-otp", auth.AuthModeGuard(cfg.AuthMode, "phone"))
		r.Post("/auth/phone/verify-otp", auth.AuthModeGuard(cfg.AuthMode, "phone"))
		r.Get("/auth/microsoft/start", auth.AuthModeGuard(cfg.AuthMode, "microsoft"))
		r.Get("/auth/microsoft/callback", auth.AuthModeGuard(cfg.AuthMode, "microsoft"))
		r.Get("/credits/balance", creditBalanceHandler(a, c))
		r.Post("/assets/upload-url", uploadURLHandler(a, s))
		r.Post("/generations/pixart", createGenerationHandler("pixel_portrait", a, c, s, g))
		r.Post("/generations/comic", createGenerationHandler("comic", a, c, s, g))
	})
	return r
}
func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data, RequestID: ""})
}
func requestLogger(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			next.ServeHTTP(w, r)
			l.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
		})
	}
}
