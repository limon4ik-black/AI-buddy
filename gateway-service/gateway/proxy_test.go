package gateway

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/security"
)

func TestGatewayBoundary(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	k := &security.Keys{Private: priv, Public: pub, Now: time.Now}
	upstream := httptest.NewServer(httpapi.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-ID") != "" || r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Role") != "" || r.Header.Get("X-Service-Token") != "server-secret" {
			t.Error("Untrusted identity forwarded")
		}
		httpapi.JSON(w, 200, map[string]string{"path": r.URL.Path})
	})))
	defer upstream.Close()
	g := Config{Keys: k, Token: "server-secret", Origin: "http://localhost:5173", AuthURL: upstream.URL, EmployeeURL: upstream.URL, ChatURL: upstream.URL}
	handler, e := New(g)
	if e != nil {
		t.Fatal(e)
	}
	s := httptest.NewServer(handler)
	defer s.Close()
	token := k.Sign("u", "u@example.test")
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/me", "/api/v1/chats", "/api/v1/services", "/api/v1/teams/backend"} {
		req(t, s.Client(), s.URL, "GET", path, token, nil, 200, "X-User-ID", "attacker", "X-Service-Token", "attacker", "X-Role", "admin", "X-Forwarded-For", "8.8.8.8", "Origin", "http://localhost:5173")
	}
	req(t, s.Client(), s.URL, "GET", "/api/v1/profile", "", nil, 401)
	req(t, s.Client(), s.URL, "GET", "/api/v1/nope", token, nil, 404)
	req(t, s.Client(), s.URL, "POST", "/api/v1/auth/refresh", "", nil, 403, "Origin", "http://evil.example")
	req(t, s.Client(), s.URL, "OPTIONS", "/api/v1/profile", "", nil, 204, "Origin", "http://localhost:5173")
	req(t, s.Client(), s.URL, "GET", "/healthz", "", nil, 200)
	g.AuthURL = "not a URL"
	if _, e = New(g); e == nil {
		t.Fatal("Invalid upstream accepted")
	}
	g.AuthURL = "http://127.0.0.1:1"
	handler, _ = New(g)
	broken := httptest.NewServer(handler)
	defer broken.Close()
	req(t, broken.Client(), broken.URL, "POST", "/api/v1/auth/login", "", nil, 502)
}
