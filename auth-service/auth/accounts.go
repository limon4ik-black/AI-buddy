package auth

import (
	"context"
	"net/http"

	"github.com/limon4ik-black/AI-buddy/shared/demodata"
	"golang.org/x/crypto/bcrypt"

	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/postgres"
)

func (a *Service) register(w http.ResponseWriter, r *http.Request) error {
	c, e := a.readCredentials(r)
	if e != nil {
		return e
	}
	// Реальный аккаунт с такой почтой уже есть - дубли не плодим.
	var exists string
	if err := a.DB.QueryRow(r.Context(), "SELECT id FROM users WHERE email=$1", c.Email).Scan(&exists); err == nil {
		return httpapi.Problem(409, "email_taken", "Аккаунт с этой почтой уже зарегистрирован — войдите по паролю")
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if e != nil {
		return e
	}
	// Аккаунт создаётся только после ввода кода: сначала заявка на почту.
	if _, e = a.DB.Exec(r.Context(), "INSERT INTO pending_registrations(email,password_hash,code_hash,code_expires_at) VALUES($1,$2,'',now()) "+
		"ON CONFLICT(email) DO UPDATE SET password_hash=EXCLUDED.password_hash, code_hash='', code_expires_at=now()",
		c.Email, string(hash)); e != nil {
		return e
	}
	a.sendPendingCodeAsync(w, r, c.Email)
	httpapi.JSON(w, http.StatusCreated, map[string]any{"message": "Код для входа отправлен на " + c.Email})
	return nil
}

func (a *Service) login(w http.ResponseWriter, r *http.Request) error {
	c, e := a.readCredentials(r)
	if e != nil {
		return e
	}
	u, e := postgres.One[Account](r.Context(), a.DB, "SELECT id,email,password_hash FROM users WHERE email=$1", c.Email)
	if e == nil {
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(c.Password)) != nil {
			return httpapi.Problem(401, "invalid_credentials", "Неверный email или пароль")
		}
		a.sendLoginCodeAsync(w, r, u.ID, u.Email)
		httpapi.JSON(w, http.StatusOK, map[string]any{"message": "Код для входа отправлен на " + u.Email})
		return nil
	}
	// Юзера нет - может, заявка ещё не подтверждена кодом.
	p, pe := postgres.One[PendingRegistration](r.Context(), a.DB, "SELECT email,password_hash FROM pending_registrations WHERE email=$1", c.Email)
	if pe != nil || bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(c.Password)) != nil {
		return httpapi.Problem(401, "invalid_credentials", "Неверный email или пароль")
	}
	a.sendPendingCodeAsync(w, r, p.Email)
	httpapi.JSON(w, http.StatusOK, map[string]any{"message": "Код для входа отправлен на " + p.Email})
	return nil
}

func (a *Service) SeedDemo(ctx context.Context) error {
	for _, intern := range demodata.Interns() {
		password := demodata.Password
		if intern.ID == "demo-intern" {
			password = "DemoPassword123!"
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err = a.DB.Exec(ctx, "INSERT INTO users(id,email,password_hash) VALUES($1,$2,$3) ON CONFLICT(email) DO NOTHING", intern.ID, intern.Email, string(hash)); err != nil {
			return err
		}
	}
	return nil
}
