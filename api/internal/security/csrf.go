package security
import("crypto/subtle";"net/http")
// RequireCSRF rejects cookie-authenticated mutations without a matching token.
func RequireCSRF(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.Method==http.MethodGet||r.Method==http.MethodHead||r.Method==http.MethodOptions{next.ServeHTTP(w,r);return};cookie,err:=r.Cookie("ai_day_csrf");if err!=nil||subtle.ConstantTimeCompare([]byte(cookie.Value),[]byte(r.Header.Get("X-CSRF-Token")))!=1{http.Error(w,"csrf validation failed",http.StatusForbidden);return};next.ServeHTTP(w,r)})}
