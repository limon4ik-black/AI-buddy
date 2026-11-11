package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
)

const loginSIDCookie = "buddy_login_sid"

func (a *Service) sendLoginCode(w http.ResponseWriter, r *http.Request, userID, email string) error {
	if a.Mailer == nil {
		return nil
	}
	sid := idgen.New() + idgen.New()
	http.SetCookie(w, &http.Cookie{
		Name: loginSIDCookie, Value: sid, Path: "/api/v1/auth", MaxAge: 600,
		HttpOnly: true, Secure: a.SecureCookie, SameSite: http.SameSiteLaxMode,
	})
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	if _, err := a.DB.Exec(r.Context(), "INSERT INTO login_confirmations(token_hash,user_id,expires_at) VALUES($1,$2,$3)",
		security.TokenHash(code+"|"+sid), userID, time.Now().Add(7*time.Minute)); err != nil {
		return err
	}
	if err := a.Mailer.Send(email, "Код для входа в AI-buddy", loginCodeText(code), loginCodeHTML(code)); err != nil {
		return err
	}
	if _, err := a.DB.Exec(r.Context(), "DELETE FROM login_confirmations WHERE user_id=$1 AND token_hash<>$2",
		userID, security.TokenHash(code+"|"+sid)); err != nil {
		return err
	}
	return nil
}

func (a *Service) sendLoginCodeAsync(w http.ResponseWriter, r *http.Request, userID, email string) {
	sid := idgen.New() + idgen.New()
	http.SetCookie(w, &http.Cookie{
		Name: loginSIDCookie, Value: sid, Path: "/api/v1/auth", MaxAge: 600,
		HttpOnly: true, Secure: a.SecureCookie, SameSite: http.SameSiteLaxMode,
	})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("login code sender panicked", "panic", r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if err := a.sendLoginCodeBody(ctx, userID, email, sid); err != nil {
			slog.Warn("login code email failed", "error", err)
		}
	}()
}

func (a *Service) sendLoginCodeBody(ctx context.Context, userID, email, sid string) error {
	if a.Mailer == nil {
		return nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	if _, err := a.DB.Exec(ctx, "INSERT INTO login_confirmations(token_hash,user_id,expires_at) VALUES($1,$2,$3)",
		security.TokenHash(code+"|"+sid), userID, time.Now().Add(7*time.Minute)); err != nil {
		return err
	}
	if err := a.Mailer.Send(email, "Код для входа в AI-buddy", loginCodeText(code), loginCodeHTML(code)); err != nil {
		return err
	}
	if _, err := a.DB.Exec(ctx, "DELETE FROM login_confirmations WHERE user_id=$1 AND token_hash<>$2",
		userID, security.TokenHash(code+"|"+sid)); err != nil {
		return err
	}
	return nil
}

func (a *Service) confirmCode(w http.ResponseWriter, r *http.Request) error {
	req, err := httpapi.Decode[struct {
		Code string `json:"code"`
	}](r)
	if err != nil {
		return err
	}
	if req.Code == "" {
		return httpapi.Invalid("Укажите код из письма")
	}
	sidCookie, err := r.Cookie(loginSIDCookie)
	if err != nil || sidCookie.Value == "" {
		return httpapi.Problem(400, "no_pending_login", "Начните вход заново")
	}

	var userID string
	err = a.DB.QueryRow(r.Context(), "DELETE FROM login_confirmations WHERE token_hash=$1 AND expires_at>now() RETURNING user_id",
		security.TokenHash(req.Code+"|"+sidCookie.Value)).Scan(&userID)
	if err != nil {
		return httpapi.Problem(400, "invalid_code", "Код неверный или устарел")
	}

	var email string
	if err = a.DB.QueryRow(r.Context(), "SELECT email FROM users WHERE id=$1", userID).Scan(&email); err != nil {
		return err
	}

	if err = a.startSession(w, r, userID); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name: loginSIDCookie, Value: "", Path: "/api/v1/auth", MaxAge: -1,
		HttpOnly: true, Secure: a.SecureCookie, SameSite: http.SameSiteLaxMode,
	})
	httpapi.JSON(w, 200, map[string]any{
		"accessToken": a.Keys.Sign(userID, email),
		"expiresIn":   900,
		"user":        map[string]string{"id": userID, "email": email, "role": "intern"},
	})
	return nil
}
