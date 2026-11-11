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

// lastToken extracts the confirmation token from the last sent letter.
func (m *captureMailer) lastToken(t *testing.T) string {
	t.Helper()
	if len(m.bodies) == 0 {
		t.Fatal("no confirmation letters sent")
	}
	line := m.bodies[len(m.bodies)-1]
	const marker = "token="
	i := strings.Index(line, marker)
	if i < 0 {
		t.Fatalf("no token in letter: %s", line)
	}
	token := line[i+len(marker):]
	if j := strings.IndexAny(token, "\r\n 	"); j >= 0 {
		token = token[:j]
	}
	return token
}

func TestEmailVerification(t *testing.T) {
	h := setup(t)
	mailer := &captureMailer{}
	h.auth.Mailer = mailer
	srv := httptest.NewServer(h.auth.Handler())
	t.Cleanup(srv.Close)

	email := "verify-" + idgen.New() + "@example.test"
	register(t, h.client, srv.URL, email, "StrongPassword123!")

	// Логин до подтверждения: 403 и повторная отправка ссылки.
	body := post(t, h.client, srv.URL, "/api/v1/auth/login", "", map[string]string{"email": email, "password": "StrongPassword123!"}, 403)
	if code(body) != "email_not_verified" {
		t.Fatalf("expected email_not_verified, got %v", body)
	}
	mailer.lastToken(t)
	if len(mailer.bodies) < 2 {
		t.Fatalf("login should have resent the link, letters: %d", len(mailer.bodies))
	}
	token := mailer.lastToken(t)

	// Подтверждение по токену из письма.
	get(t, h.client, srv.URL, "/api/v1/auth/confirm?token="+token, "", 200)

	// Повторное подтверждение тем же токеном: 400, токен одноразовый.
	get(t, h.client, srv.URL, "/api/v1/auth/confirm?token="+token, "", 400)

	// После подтверждения логин проходит и отдает accessToken.
	ok := post(t, h.client, srv.URL, "/api/v1/auth/login", "", map[string]string{"email": email, "password": "StrongPassword123!"}, 200)
	if access, ok := ok["accessToken"].(string); !ok || access == "" {
		t.Fatalf("no accessToken after confirmation: %v", ok)
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
