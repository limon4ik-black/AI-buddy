package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
)

// captureMailer stores letter bodies instead of sending them.
type captureMailer struct {
	bodies []string
}

func (m *captureMailer) Send(to, subject, body string) error {
	m.bodies = append(m.bodies, body)
	return nil
}

// lastToken extracts the login token from the last sent letter.
func (m *captureMailer) lastToken(t *testing.T) string {
	t.Helper()
	if len(m.bodies) == 0 {
		t.Fatal("no login letters sent")
	}
	line := m.bodies[len(m.bodies)-1]
	const marker = "token="
	i := strings.Index(line, marker)
	if i < 0 {
		t.Fatalf("no token in letter: %s", line)
	}
	token := line[i+len(marker):]
	if j := strings.IndexAny(token, "\r\n \t"); j >= 0 {
		token = token[:j]
	}
	return token
}

// noRedirectAtRoot follows redirects but stops on the final redirect to the
// frontend root made by the confirmation endpoint.
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
	if len(mailer.bodies) != 1 {
		t.Fatalf("register should send exactly one letter, got %d", len(mailer.bodies))
	}

	// Неверный пароль: 401, письмо не отправляется.
	wrong := post(t, h.client, srv.URL, "/api/v1/auth/login", "", map[string]string{"email": email, "password": "WrongPassword123!"}, 401)
	if code(wrong) != "invalid_credentials" {
		t.Fatalf("expected invalid_credentials, got %v", wrong)
	}
	if len(mailer.bodies) != 1 {
		t.Fatalf("wrong password must not send letters, got %d", len(mailer.bodies))
	}

	// Верный пароль: 200 + письмо со ссылкой; сессии пока нет.
	post(t, h.client, srv.URL, "/api/v1/auth/login", "", map[string]string{"email": email, "password": "StrongPassword123!"}, 200)
	token := mailer.lastToken(t)

	// До перехода по ссылке cookie-сессии нет: refresh 401.
	req(t, noRedirectAtRoot(h.client), srv.URL, "POST", "/api/v1/auth/refresh", "", nil, 401)

	// Переход по ссылке: 302 на фронт, cookie-сессия выдана.
	resp, err := noRedirectAtRoot(h.client).Get(srv.URL + "/api/v1/auth/confirm?token=" + token)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("confirm status: %d", resp.StatusCode)
	}

	// Повторное использование токена: 400.
	resp2, err := noRedirectAtRoot(h.client).Get(srv.URL + "/api/v1/auth/confirm?token=" + token)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusBadRequest {
		t.Fatalf("token reuse status: %d", resp2.StatusCode)
	}

	// Сессия жива: refresh выдает accessToken, /me отвечает.
	access := refreshAccessToken(t, h.client, srv.URL)
	if id := meUserID(t, h.client, srv.URL, access); id == "" {
		t.Fatal("empty user id from /me")
	}
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
