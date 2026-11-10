package auth

import (
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
)

type Service struct {
	DB           *pgxpool.Pool
	Keys         *security.Keys
	SecureCookie bool
	mu           sync.Mutex
	attempts     map[string]attempt
}

func (a *Service) Handler() http.Handler {
	m := http.NewServeMux()
	httpapi.Health(m, a.DB, "auth")
	m.HandleFunc("POST /api/v1/auth/register", httpapi.Handle(a.limited(a.register)))
	m.HandleFunc("POST /api/v1/auth/login", httpapi.Handle(a.limited(a.login)))
	m.HandleFunc("POST /api/v1/auth/refresh", httpapi.Handle(a.refresh))
	m.HandleFunc("POST /api/v1/auth/logout", httpapi.Handle(a.logout))
	m.HandleFunc("GET /api/v1/auth/me", httpapi.Handle(httpapi.Authenticated(a.Keys, func(w http.ResponseWriter, r *http.Request) error {
		c := httpapi.User(r)
		httpapi.JSON(w, 200, map[string]string{"id": c.Subject, "email": c.Email, "role": c.Role})
		return nil
	})))
	return httpapi.Middleware(m)
}
