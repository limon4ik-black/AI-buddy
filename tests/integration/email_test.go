package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/limon4ik-black/AI-buddy/shared/platform/idgen"
)

type captureMailer struct {
	mu     sync.Mutex
	bodies []string
}

func (m *captureMailer) Send(to, subject, text, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bodies = append(m.bodies, text)
	return nil
}

func (m *captureMailer) letters() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.bodies...)
}

func waitForLetter(t *testing.T, m *captureMailer, want int) {
	t.Helper()
	for i := 0; i < 100; i++ {
		if len(m.letters()) >= want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("letters not delivered: have %d, want %d", len(m.letters()), want)
}

func (m *captureMailer) lastCode(t *testing.T) string {
	t.Helper()
	if len(m.letters()) == 0 {
		t.Fatal("no login letters sent")
	}
	line := m.letters()[len(m.letters())-1]
	const marker = "Код для входа в AI-buddy: "
	i := strings.Index(line, marker)
	if i < 0 {
		t.Fatalf("no code in letter: %s", line)
	}
	return line[i+len(marker):][:6]
}

func noRedirectAtRoot(client *http.Client) *http.Client {
	return &http.Client{
		Jar:     client.Jar,
		Timeout: client.Timeout,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if req.URL.Path == "/" && req.URL.RawQuery == "" {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func TestEmailLoginFlow(t *testing.T) {
	h := setup(t)
	mailer := &captureMailer{}
	h.auth.Mailer = mailer
	srv := httptest.NewServer(h.auth.Handler())
	t.Cleanup(srv.Close)

	email := "verify-" + idgen.New() + "@example.test"
	register(t, h.client, srv.URL, email, "StrongPassword123!")
	waitForLetter(t, mailer, 1)
	if len(mailer.letters()) != 1 {
		t.Fatalf("register should send exactly one letter, got %d", len(mailer.letters()))
	}
	// Аккаунт не должен существовать до ввода кода.
	var count int
	if e := h.auth.DB.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE email=$1", email).Scan(&count); e != nil || count != 0 {
		t.Fatalf("user must not exist before confirmation: count=%d err=%v", count, e)
	}

	wrong := post(t, h.client, srv.URL, "/api/v1/auth/login", "", map[string]string{"email": email, "password": "WrongPassword123!"}, 401)
	if code(wrong) != "invalid_credentials" {
		t.Fatalf("expected invalid_credentials, got %v", wrong)
	}
	if len(mailer.letters()) != 1 {
		t.Fatalf("wrong password must not send letters, got %d", len(mailer.letters()))
	}

	post(t, h.client, srv.URL, "/api/v1/auth/login", "", map[string]string{"email": email, "password": "StrongPassword123!"}, 200)
	waitForLetter(t, mailer, 2)
	codeValue := mailer.lastCode(t)
	req(t, h.client, srv.URL, "POST", "/api/v1/auth/refresh", "", nil, 401)

	bad := post(t, h.client, srv.URL, "/api/v1/auth/confirm", "", map[string]string{"code": "000000"}, 400)
	if code(bad) != "invalid_code" {
		t.Fatalf("expected invalid_code, got %v", bad)
	}

	alien := post(t, &http.Client{}, srv.URL, "/api/v1/auth/confirm", "", map[string]string{"code": codeValue}, 400)
	if code(alien) != "no_pending_login" {
		t.Fatalf("expected no_pending_login for foreign window, got %v", alien)
	}

	ok := post(t, h.client, srv.URL, "/api/v1/auth/confirm", "", map[string]string{"code": codeValue}, 200)
	access, has := ok["accessToken"].(string)
	if !has || access == "" {
		t.Fatalf("no accessToken after confirm: %v (sent code %q, letters %q)", ok, codeValue, mailer.letters())
	}
	if id := meUserID(t, h.client, srv.URL, access); id == "" {
		t.Fatal("empty user id from /me")
	}
	refreshAccessToken(t, h.client, srv.URL)

	req(t, &http.Client{}, srv.URL, "POST", "/api/v1/auth/confirm", "", map[string]string{"code": "123456"}, 400)
}

func register(t *testing.T, client *http.Client, base, email, password string) {
	t.Helper()
	post(t, client, base, "/api/v1/auth/register", "", map[string]string{"email": email, "password": password}, 201)
}

func post(t *testing.T, client *http.Client, base, path, token string, body any, status int) map[string]any {
	t.Helper()
	raw := req(t, client, base, "POST", path, token, body, status)
	var out map[string]any
	if status != 204 && len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("bad json from %s: %s", path, raw)
		}
	}
	return out
}

func get(t *testing.T, client *http.Client, base, path, token string, status int) map[string]any {
	t.Helper()
	raw := req(t, client, base, "GET", path, token, nil, status)
	var out map[string]any
	if status != 204 && len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out
}

func code(m map[string]any) string {
	err, _ := m["error"].(map[string]any)
	c, _ := err["code"].(string)
	return c
}
