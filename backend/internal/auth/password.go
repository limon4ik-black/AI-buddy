package auth

import (
	"net/http"
	"strings"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
	"golang.org/x/crypto/bcrypt"
)

// passwordReset mails a code for the account if it exists; the answer is
// always the same so the endpoint cannot be used to enumerate accounts.
func (a *Service) passwordReset(w http.ResponseWriter, r *http.Request) error {
	req, err := httpapi.Decode[struct {
		Email string `json:"email"`
	}](r)
	if err != nil {
		return err
	}
	email, err := normalizeEmail(req.Email)
	if err != nil {
		return err
	}
	var id string
	if e := a.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE email=$1", email).Scan(&id); e == nil {
		a.sendResetCodeAsync(w, r, id, email)
	}
	httpapi.JSON(w, 200, map[string]any{"message": "Если аккаунт существует, код отправлен на почту"})
	return nil
}

// passwordConfirm swaps the password after the reset code, opens a session
// and kills every other refresh session of the account.
func (a *Service) passwordConfirm(w http.ResponseWriter, r *http.Request) error {
	req, err := httpapi.Decode[struct {
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}](r)
	if err != nil {
		return err
	}
	if len(req.NewPassword) < 10 || len(req.NewPassword) > 72 {
		return httpapi.Invalid("Пароль должен содержать от 10 до 72 байт")
	}
	sidCookie, err := r.Cookie(loginSIDCookie)
	if err != nil || sidCookie.Value == "" {
		return httpapi.Problem(400, "no_pending_login", "Начните восстановление заново")
	}
	var userID, email string
	err = a.DB.QueryRow(r.Context(), "DELETE FROM login_confirmations WHERE token_hash=$1 AND purpose='reset' AND expires_at>now() RETURNING user_id",
		security.TokenHash(req.Code+"|"+sidCookie.Value)).Scan(&userID)
	if err != nil {
		return httpapi.Problem(400, "invalid_code", "Код неверный или устарел")
	}
	if err = a.DB.QueryRow(r.Context(), "SELECT email FROM users WHERE id=$1", userID).Scan(&email); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err = a.DB.Exec(r.Context(), "UPDATE users SET password_hash=$1 WHERE id=$2", string(hash), userID); err != nil {
		return err
	}
	keep := ""
	if c, e := r.Cookie("buddy_refresh"); e == nil {
		keep = security.TokenHash(c.Value)
	}
	if _, err = a.DB.Exec(r.Context(), "DELETE FROM refresh_sessions WHERE user_id=$1 AND token_hash<>$2", userID, keep); err != nil {
		return err
	}
	if err = a.session(w, r, userID, email, 200); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name: loginSIDCookie, Value: "", Path: "/api/v1/auth", MaxAge: -1,
		HttpOnly: true, Secure: a.SecureCookie, SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	return email, validEmail(email)
}
