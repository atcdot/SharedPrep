package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/atcdot/SharedPrep/internal/auth"
	"github.com/atcdot/SharedPrep/internal/storage"
)

type AuthHandler struct {
	db       *storage.Postgres
	verifier *auth.Verifier
	secret   []byte
	logger   *slog.Logger
}

func NewAuthHandler(db *storage.Postgres, verifier *auth.Verifier, secret []byte, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{db: db, verifier: verifier, secret: secret, logger: logger}
}

func (h *AuthHandler) TelegramLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if body.IDToken == "" {
		http.Error(w, "missing id_token", http.StatusBadRequest)
		return
	}

	user, err := h.verifier.VerifyIDToken(r.Context(), body.IDToken)
	if err != nil {
		h.logger.Warn("telegram id_token verification failed", "error", err)
		http.Error(w, "invalid auth", http.StatusUnauthorized)
		return
	}

	firstName, lastName := auth.SplitName(user.Name)

	var userID string
	err = h.db.Pool.QueryRow(r.Context(), `
		INSERT INTO users (telegram_id, username, first_name, last_name, photo_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (telegram_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			photo_url = EXCLUDED.photo_url,
			updated_at = now()
		RETURNING id
	`, user.TelegramID, user.Username, firstName, lastName, user.PhotoURL).Scan(&userID)
	if err != nil {
		h.logger.Error("failed to upsert user", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(userID, h.secret)
	if err != nil {
		h.logger.Error("failed to generate token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":         userID,
		"telegramId": user.TelegramID,
		"username":   user.Username,
		"firstName":  firstName,
		"lastName":   lastName,
		"photoUrl":   user.PhotoURL,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseToken(cookie.Value, h.secret)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var username, firstName, lastName, photoURL string
	var telegramID int64
	var displayName *string
	var showTelegram bool
	err = h.db.Pool.QueryRow(r.Context(), `
		SELECT telegram_id, username, first_name, last_name, photo_url, display_name, show_telegram
		FROM users WHERE id = $1
	`, claims.UserID).Scan(&telegramID, &username, &firstName, &lastName, &photoURL, &displayName, &showTelegram)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":           claims.UserID,
		"telegramId":   telegramID,
		"username":     username,
		"firstName":    firstName,
		"lastName":     lastName,
		"photoUrl":     photoURL,
		"displayName":  displayName,
		"showTelegram": showTelegram,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusOK)
}
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseToken(cookie.Value, h.secret)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		DisplayName  *string `json:"display_name"`
		ShowTelegram *bool   `json:"show_telegram"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if body.DisplayName != nil {
		_, err = h.db.Pool.Exec(r.Context(), `
			UPDATE users SET display_name = $1, updated_at = now() WHERE id = $2
		`, *body.DisplayName, claims.UserID)
		if err != nil {
			h.logger.Error("failed to update profile", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	if body.ShowTelegram != nil {
		_, err = h.db.Pool.Exec(r.Context(), `
			UPDATE users SET show_telegram = $1, updated_at = now() WHERE id = $2
		`, *body.ShowTelegram, claims.UserID)
		if err != nil {
			h.logger.Error("failed to update show_telegram", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"displayName":  body.DisplayName,
		"showTelegram": body.ShowTelegram,
	})
}

