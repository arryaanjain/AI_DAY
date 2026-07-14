package httpapi

import (
	"net/http"
	"strings"
)

func cors(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins)*2)
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimRight(origin, "/")
		allowed[trimmed] = struct{}{}
		// Map localhost <-> 127.0.0.1 aliases for seamless dev experience
		if strings.Contains(trimmed, "://localhost:") {
			allowed[strings.Replace(trimmed, "://localhost:", "://127.0.0.1:", 1)] = struct{}{}
		} else if strings.Contains(trimmed, "://127.0.0.1:") {
			allowed[strings.Replace(trimmed, "://127.0.0.1:", "://localhost:", 1)] = struct{}{}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimRight(r.Header.Get("Origin"), "/")
			if origin != "" {
				if _, ok := allowed[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token, X-CSRF-Token, Idempotency-Key, Accept, Origin, X-Requested-With")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
