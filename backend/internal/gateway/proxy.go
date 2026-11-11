package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
)

type Config struct {
	Keys                                     *security.Keys
	Token, Origin, AuthURL, InfoURL, ChatURL string
}

func New(cfg Config) (http.Handler, error) {
	g := &cfg
	m := http.NewServeMux()
	httpapi.Health(m, nil, "gateway")
	proxies := map[string]http.Handler{}
	for name, target := range map[string]string{"auth": g.AuthURL, "info": g.InfoURL, "chat": g.ChatURL} {
		u, e := url.Parse(target)
		if e != nil || u.Scheme != "http" || u.Host == "" {
			return nil, httpapi.Invalid("Некорректный URL " + name)
		}
		p := &httputil.ReverseProxy{Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(u)
			pr.Out.Header.Del("X-User-ID")
			pr.Out.Header.Del("X-Role")
			pr.Out.Header.Del("Forwarded")
			pr.Out.Header.Del("X-Forwarded-For")
			pr.Out.Header.Del("X-Forwarded-Host")
			pr.Out.Header.Del("X-Forwarded-Proto")
			pr.Out.Header.Set("X-Service-Token", g.Token)
		}, Transport: &http.Transport{ResponseHeaderTimeout: 40 * time.Second}, ErrorHandler: func(w http.ResponseWriter, r *http.Request, e error) {
			httpapi.JSON(w, 502, map[string]any{"error": map[string]any{"code": "upstream_unavailable", "message": "Сервис временно недоступен", "requestId": httpapi.RequestID(r.Context())}})
		}}
		proxies[name] = p
	}
	m.Handle("/api/v1/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin != g.Origin {
			httpapi.Handle(func(http.ResponseWriter, *http.Request) error {
				return httpapi.Problem(403, "origin_denied", "Недопустимый origin")
			})(w, r)
			return
		}
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.WriteHeader(204)
			return
		}
		path := r.URL.Path
		name := ""
		public := false
		switch {
		case path == "/api/v1/auth/register" || path == "/api/v1/auth/login" || path == "/api/v1/auth/refresh" || path == "/api/v1/auth/logout" || path == "/api/v1/auth/google/login" || path == "/api/v1/auth/google/callback" || path == "/api/v1/auth/confirm":
			name = "auth"
			public = true
		case path == "/api/v1/auth/me":
			name = "auth"
		case path == "/api/v1/chats" || strings.HasPrefix(path, "/api/v1/chats/"):
			name = "chat"
		case path == "/api/v1/profile" || path == "/api/v1/checklist" || strings.HasPrefix(path, "/api/v1/checklist/tasks/") || path == "/api/v1/teams" || strings.HasPrefix(path, "/api/v1/teams/") || path == "/api/v1/services" || strings.HasPrefix(path, "/api/v1/services/") || path == "/api/v1/tools" || path == "/api/v1/information" || strings.HasPrefix(path, "/api/v1/information/") || path == "/api/v1/access-requests":
			name = "info"
		default:
			httpapi.Handle(func(http.ResponseWriter, *http.Request) error { return httpapi.NotFound() })(w, r)
			return
		}
		forward := func(w http.ResponseWriter, r *http.Request) error { proxies[name].ServeHTTP(w, r); return nil }
		if public {
			_ = forward(w, r)
		} else {
			httpapi.Handle(httpapi.Authenticated(g.Keys, forward))(w, r)
		}
	}))
	return httpapi.Middleware(m), nil
}
