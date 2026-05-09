package api

import (
	"log/slog"
	"net/http"

	"github.com/atcdot/SharedPrep/internal/auth"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type AuthMiddleware struct {
	jwtSecret []byte
	logger    *slog.Logger
}

func NewAuthMiddleware(jwtSecret []byte, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret, logger: logger}
}

func (m *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session_token"); err == nil {
			claims, err := auth.ParseToken(cookie.Value, m.jwtSecret)
			if err == nil {
				r = r.WithContext(WithUserID(r.Context(), claims.UserID))
			}
		}
		next.ServeHTTP(w, r)
	})
}
