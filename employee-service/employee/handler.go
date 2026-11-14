package employee

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/security"
)

type Service struct {
	DB                                                        *pgxpool.Pool
	Keys                                                      *security.Keys
	GatewayToken, ChatToken, IndexerToken, CorpToken, CorpURL string
	Client                                                    *http.Client
}

func (i *Service) Handler() http.Handler {
	m := http.NewServeMux()
	httpapi.Health(m, i.DB, "employee")
	public := func(pattern string, fn httpapi.Endpoint) {
		m.HandleFunc(pattern, httpapi.Handle(httpapi.ServiceToken(i.GatewayToken, httpapi.Authenticated(i.Keys, i.withProfile(fn)))))
	}
	public("GET /api/v1/profile", i.profile)
	public("PATCH /api/v1/profile", i.updateProfile)
	public("GET /api/v1/teams", i.teams)
	public("GET /api/v1/teams/{id}", i.team)
	public("GET /api/v1/services", i.services)
	public("GET /api/v1/services/{id}", i.technicalService)
	public("GET /api/v1/tools", i.tools)
	public("GET /api/v1/access-requests", i.requests)
	public("POST /api/v1/access-requests", i.createRequest)
	m.HandleFunc("GET /internal/v1/context", httpapi.Handle(httpapi.ServiceToken(i.ChatToken, httpapi.Authenticated(i.Keys, i.withProfile(i.context)))))
	m.HandleFunc("GET /internal/v1/users/{userId}/context", httpapi.Handle(httpapi.ServiceToken(i.ChatToken, httpapi.Authenticated(i.Keys, func(w http.ResponseWriter, r *http.Request) error {
		if r.PathValue("userId") != httpapi.User(r).Subject {
			return httpapi.Problem(http.StatusForbidden, "identity_mismatch", "Нельзя запрашивать контекст другого пользователя")
		}
		return i.withProfile(i.context)(w, r)
	}))))
	for pattern, fn := range map[string]httpapi.Endpoint{
		"GET /internal/v1/teams":         i.teams,
		"GET /internal/v1/teams/{id}":    i.team,
		"GET /internal/v1/services":      i.services,
		"GET /internal/v1/services/{id}": i.technicalService,
	} {
		m.HandleFunc(pattern, httpapi.Handle(httpapi.ServiceToken(i.IndexerToken, fn)))
	}
	return httpapi.Middleware(m)
}

func (i *Service) withProfile(fn httpapi.Endpoint) httpapi.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) error {
		if e := i.ensureProfile(r); e != nil {
			return e
		}
		return fn(w, r)
	}
}
