package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
)

// sendLoginLink rotates the login-confirmation token and mails the link.
// Mailer == nil means sending is disabled (nothing happens).
// Каждый вход требует свежей ссылки: сам пароль сессию не открывает.
func (a *Service) sendLoginLink(ctx context.Context, userID, email string) error {
	if a.Mailer == nil {
		return nil
	}
	token := idgen.New() + idgen.New()
	if _, err := a.DB.Exec(ctx, "DELETE FROM login_confirmations WHERE user_id=$1", userID); err != nil {
		return err
	}
	if _, err := a.DB.Exec(ctx, "INSERT INTO login_confirmations(token_hash,user_id,expires_at) VALUES($1,$2,$3)",
		security.TokenHash(token), userID, time.Now().Add(15*time.Minute)); err != nil {
		return err
	}
	origin := strings.TrimRight(config.Env("APP_ORIGIN", "http://localhost:8080"), "/")
	return a.Mailer.Send(email, "Вход в AI-buddy", loginLinkText(origin, token))
}

// confirm consumes the token from the login letter and opens the session:
// only the inbox owner can complete the login.
func (a *Service) confirm(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if token == "" {
		return httpapi.Invalid("Отсутствует токен подтверждения")
	}
	var userID string
	err := a.DB.QueryRow(r.Context(), "DELETE FROM login_confirmations WHERE token_hash=$1 AND expires_at>now() RETURNING user_id",
		security.TokenHash(token)).Scan(&userID)
	if err != nil {
		return httpapi.Problem(400, "invalid_confirmation", "Ссылка недействительна или устарела")
	}
	if err = a.startSession(w, r, userID); err != nil {
		return err
	}
	origin := strings.TrimRight(config.Env("APP_ORIGIN", "http://localhost:8080"), "/")
	http.Redirect(w, r, origin+"/", http.StatusFound)
	return nil
}
