package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
)

// sendConfirmation rotates the confirmation token and mails the link.
// Mailer == nil means sending is disabled (nothing happens).
func (a *Service) sendConfirmation(ctx context.Context, userID, email string) error {
	if a.Mailer == nil {
		return nil
	}
	token := idgen.New() + idgen.New()
	if _, err := a.DB.Exec(ctx, "DELETE FROM email_confirmations WHERE user_id=$1", userID); err != nil {
		return err
	}
	if _, err := a.DB.Exec(ctx, "INSERT INTO email_confirmations(token_hash,user_id,expires_at) VALUES($1,$2,$3)",
		security.TokenHash(token), userID, time.Now().Add(24*time.Hour)); err != nil {
		return err
	}
	origin := config.Env("APP_ORIGIN", "http://localhost:8080")
	return a.Mailer.Send(email, "Подтверждение почты AI-buddy", confirmationLinkText(origin, token))
}

// confirm consumes the token from the link and verifies the email.
func (a *Service) confirm(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if token == "" {
		return httpapi.Invalid("Отсутствует токен подтверждения")
	}
	tag, err := a.DB.Exec(r.Context(), "UPDATE users SET email_verified=true WHERE id=("+
		"SELECT user_id FROM email_confirmations WHERE token_hash=$1 AND expires_at>now())",
		security.TokenHash(token))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpapi.Problem(400, "invalid_confirmation", "Ссылка недействительна или устарела")
	}
	if _, err = a.DB.Exec(r.Context(), "DELETE FROM email_confirmations WHERE token_hash=$1", security.TokenHash(token)); err != nil {
		return err
	}
	httpapi.JSON(w, 200, map[string]any{"ok": true, "message": "Почта подтверждена — можно входить"})
	return nil
}
