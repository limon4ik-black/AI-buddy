package auth

import (
	"net/http"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
)

func (a *Service) cookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: "buddy_refresh", Value: value, Path: "/api/v1/auth", HttpOnly: true, Secure: a.SecureCookie, SameSite: http.SameSiteStrictMode, MaxAge: maxAge})
}

func (a *Service) session(w http.ResponseWriter, r *http.Request, id, email string, status int) error {
	if e := a.startSession(w, r, id); e != nil {
		return e
	}
	httpapi.JSON(w, status, map[string]any{"accessToken": a.Keys.Sign(id, email), "expiresIn": 900, "user": map[string]string{"id": id, "email": email, "role": "intern"}})
	return nil
}

func (a *Service) startSession(w http.ResponseWriter, r *http.Request, id string) error {
	token := idgen.New() + idgen.New()
	_, e := a.DB.Exec(r.Context(), "INSERT INTO refresh_sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)", security.TokenHash(token), id, time.Now().Add(7*24*time.Hour))
	if e != nil {
		return e
	}
	a.cookie(w, token, 7*24*3600)
	return nil
}

func (a *Service) refresh(w http.ResponseWriter, r *http.Request) error {
	c, e := r.Cookie("buddy_refresh")
	if e != nil {
		return httpapi.Problem(401, "invalid_session", "Войдите в систему")
	}
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		return e
	}
	defer tx.Rollback(r.Context())
	var id, email string
	e = tx.QueryRow(r.Context(), "DELETE FROM refresh_sessions WHERE token_hash=$1 AND expires_at>now() RETURNING user_id", security.TokenHash(c.Value)).Scan(&id)
	if e != nil {
		return httpapi.Problem(401, "invalid_session", "Сессия истекла. Войдите снова")
	}
	if e = tx.QueryRow(r.Context(), "SELECT email FROM users WHERE id=$1", id).Scan(&email); e != nil {
		return e
	}
	token := idgen.New() + idgen.New()
	if _, e = tx.Exec(r.Context(), "INSERT INTO refresh_sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)", security.TokenHash(token), id, time.Now().Add(7*24*time.Hour)); e != nil {
		return e
	}
	if e = tx.Commit(r.Context()); e != nil {
		return e
	}
	a.cookie(w, token, 7*24*3600)
	httpapi.JSON(w, 200, map[string]any{"accessToken": a.Keys.Sign(id, email), "expiresIn": 900, "user": map[string]string{"id": id, "email": email, "role": "intern"}})
	return nil
}

func (a *Service) logout(w http.ResponseWriter, r *http.Request) error {
	if c, e := r.Cookie("buddy_refresh"); e == nil {
		if _, e = a.DB.Exec(r.Context(), "DELETE FROM refresh_sessions WHERE token_hash=$1", security.TokenHash(c.Value)); e != nil {
			return e
		}
	}
	a.cookie(w, "", -1)
	httpapi.JSON(w, 204, nil)
	return nil
}
