package admin

import (
 "net/http"
 "github.com/arryaanjain/AI_DAY/internal/auth"
)
// RequireAdmin is backend authorization, never a frontend-route check.
func RequireAdmin(service *auth.Service,next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){cookie,err:=r.Cookie(auth.SessionCookieName);if err!=nil{http.Error(w,"authentication required",http.StatusUnauthorized);return};user,err:=service.CurrentUser(r.Context(),cookie.Value);if err!=nil||!user.IsAdmin{http.Error(w,"administrator access required",http.StatusForbidden);return};next.ServeHTTP(w,r)})}
