package auth

import (
	"context"
	"net/http"

	"github.com/limon4ik-black/AI-buddy/backend/internal/demodata"
	"golang.org/x/crypto/bcrypt"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (a *Service) register(w http.ResponseWriter, r *http.Request) error {
	c, e := a.readCredentials(r)
	if e != nil {
		return e
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if e != nil {
		return e
	}
	id := idgen.New()
	_, e = a.DB.Exec(r.Context(), "INSERT INTO users(id,email,password_hash) VALUES($1,$2,$3)", id, c.Email, string(hash))
	if e != nil {
		return e
	}
	return a.session(w, r, id, c.Email, 201)
}

func (a *Service) login(w http.ResponseWriter, r *http.Request) error {
	c, e := a.readCredentials(r)
	if e != nil {
		return e
	}
	u, e := postgres.One[Account](r.Context(), a.DB, "SELECT id,email,password_hash FROM users WHERE email=$1", c.Email)
	if e != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(c.Password)) != nil {
		return httpapi.Problem(401, "invalid_credentials", "Неверный email или пароль")
	}
	return a.session(w, r, u.ID, u.Email, 200)
}

// SeedDemo creates synthetic test accounts without replacing existing passwords.
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
