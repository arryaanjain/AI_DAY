package httpapi

import (
	"net/http"

	"github.com/arryaanjain/AI_DAY/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Admin handlers for dashboard and user management

func adminMiddleware(a *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := authenticatedUser(w, r, a)
			if !ok {
				return
			}
			if !user.IsAdmin {
				errorResponse(w, http.StatusForbidden, "ADMIN_REQUIRED", "This action requires administrator privileges.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func adminUsersListHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(), `
			SELECT id::text, COALESCE(email::text, ''), COALESCE(phone_e164, ''), COALESCE(display_name, ''), status, created_at
			FROM users
			ORDER BY created_at DESC
			LIMIT 100
		`)
		if err != nil {
			errorResponse(w, http.StatusServiceUnavailable, "QUERY_ERROR", "Failed to fetch users.")
			return
		}
		defer rows.Close()

		type UserRow struct {
			ID          string `json:"id"`
			Email       string `json:"email"`
			Phone       string `json:"phone"`
			DisplayName string `json:"displayName"`
			Status      string `json:"status"`
			CreatedAt   string `json:"createdAt"`
		}

		var users []UserRow
		for rows.Next() {
			var u UserRow
			if err := rows.Scan(&u.ID, &u.Email, &u.Phone, &u.DisplayName, &u.Status, &u.CreatedAt); err != nil {
				continue
			}
			users = append(users, u)
		}

		respond(w, http.StatusOK, map[string]any{"users": users, "total": len(users)})
	}
}

func adminGenerationsListHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(), `
			SELECT id::text, user_id::text, module, status, created_at
			FROM generation_jobs
			ORDER BY created_at DESC
			LIMIT 100
		`)
		if err != nil {
			errorResponse(w, http.StatusServiceUnavailable, "QUERY_ERROR", "Failed to fetch jobs.")
			return
		}
		defer rows.Close()

		type JobRow struct {
			ID        string `json:"id"`
			UserID    string `json:"userId"`
			Module    string `json:"module"`
			Status    string `json:"status"`
			CreatedAt string `json:"createdAt"`
		}

		var jobs []JobRow
		for rows.Next() {
			var j JobRow
			if err := rows.Scan(&j.ID, &j.UserID, &j.Module, &j.Status, &j.CreatedAt); err != nil {
				continue
			}
			jobs = append(jobs, j)
		}

		respond(w, http.StatusOK, map[string]any{"jobs": jobs, "total": len(jobs)})
	}
}

func adminStatsHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userCount, completedJobs, totalAssets int64

		_ = db.QueryRow(r.Context(), `SELECT COUNT(*) FROM users`).Scan(&userCount)
		_ = db.QueryRow(r.Context(), `SELECT COUNT(*) FROM generation_jobs WHERE status = 'completed'`).Scan(&completedJobs)
		_ = db.QueryRow(r.Context(), `SELECT COUNT(*) FROM assets`).Scan(&totalAssets)

		respond(w, http.StatusOK, map[string]any{
			"stats": map[string]any{
				"totalUsers":    userCount,
				"completedJobs": completedJobs,
				"totalAssets":   totalAssets,
			},
		})
	}
}
