package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
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
	if _, err := a.DB.Exec(ctx, "INSERT INTO login_confirmations(token_hash,user_id,purpose,expires_at) VALUES($1,$2,'login',$3)",
		security.TokenHash(code+"|"+sid), userID, time.Now().Add(7*time.Minute)); err != nil {
		return err
	}
	if err := a.Mailer.Send(email, "Код для входа в AI-buddy", loginCodeText(code), loginCodeHTML(code)); err != nil {
		return err
	}
	if _, err := a.DB.Exec(ctx, "DELETE FROM login_confirmations WHERE user_id=$1 AND purpose='login' AND token_hash<>$2",
		userID, security.TokenHash(code+"|"+sid)); err != nil {
		return err
	}
	return nil
}

// sendResetCodeAsync mails the code that proves ownership before a password change.
func (a *Service) sendResetCodeAsync(w http.ResponseWriter, r *http.Request, userID, email string) {
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
		if err := a.sendResetCode(ctx, userID, email, sid); err != nil {
			slog.Warn("login code email failed", "error", err)
		}
	}()
}

func (a *Service) sendResetCode(ctx context.Context, userID, email, sid string) error {
	if a.Mailer == nil {
		return nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	hash := security.TokenHash(code + "|" + sid)
	if _, err := a.DB.Exec(ctx, "INSERT INTO login_confirmations(token_hash,user_id,purpose,expires_at) VALUES($1,$2,'reset',$3)",
		hash, userID, time.Now().Add(7*time.Minute)); err != nil {
		return err
	}
	if err := a.Mailer.Send(email, "Код для смены пароля в AI-buddy", loginCodeText(code), loginCodeHTML(code)); err != nil {
		return err
	}
	if _, err := a.DB.Exec(ctx, "DELETE FROM login_confirmations WHERE user_id=$1 AND purpose='reset' AND token_hash<>$2", userID, hash); err != nil {
		return err
	}
	return nil
}

// sendPendingCodeAsync mails the code that completes a pending registration.
func (a *Service) sendPendingCodeAsync(w http.ResponseWriter, r *http.Request, email string) {
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
		if err := a.sendPendingCode(ctx, email, sid); err != nil {
			slog.Warn("login code email failed", "error", err)
		}
	}()
}

// sendPendingCode stores the code hash in the pending registration row.
func (a *Service) sendPendingCode(ctx context.Context, email, sid string) error {
	if a.Mailer == nil {
		return nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	if _, err := a.DB.Exec(ctx, "UPDATE pending_registrations SET code_hash=$1, code_expires_at=$2 WHERE email=$3",
		security.TokenHash(code+"|"+sid), time.Now().Add(7*time.Minute), email); err != nil {
		return err
	}
	return a.Mailer.Send(email, "Код для входа в AI-buddy", loginCodeText(code), loginCodeHTML(code))
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
	err = a.DB.QueryRow(r.Context(), "DELETE FROM login_confirmations WHERE token_hash=$1 AND purpose='login' AND expires_at>now() RETURNING user_id",
		security.TokenHash(req.Code+"|"+sidCookie.Value)).Scan(&userID)
	switch {
	case err == nil:
	case errors.Is(err, pgx.ErrNoRows):
		// Код не из сессии существующего юзера - может, это завершение регистрации.
		var email, hash string
		if e := a.DB.QueryRow(r.Context(), "DELETE FROM pending_registrations WHERE code_hash=$1 AND code_expires_at>now() RETURNING email, password_hash",
			security.TokenHash(req.Code+"|"+sidCookie.Value)).Scan(&email, &hash); e != nil {
			return httpapi.Problem(400, "invalid_code", "Код неверный или устарел")
		}
		userID = idgen.New()
		if _, e := a.DB.Exec(r.Context(), "INSERT INTO users(id,email,password_hash) VALUES($1,$2,$3)",
			userID, email, hash); e != nil {
			return e
		}
	default:
		return err
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
