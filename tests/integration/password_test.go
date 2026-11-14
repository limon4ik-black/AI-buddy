package integration

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPasswordReset(t *testing.T) {
	h := setup(t)
	mailer := &captureMailer{}
	h.auth.Mailer = mailer
	srv := httptest.NewServer(h.auth.Handler())
	t.Cleanup(srv.Close)

	// Живая сессия у владельца.
	post(t, h.client, srv.URL, "/api/v1/auth/login", "", map[string]string{"email": "intern@example.test", "password": "DemoPassword123!"}, 200)
	waitForLetter(t, mailer, 1)
	req(t, h.client, srv.URL, "POST", "/api/v1/auth/confirm", "", map[string]string{"code": mailer.lastCode(t)}, 200)

	// Сторонний клиент запрашивает сброс.
	jar, _ := cookiejar.New(nil)
	resetter := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	post(t, resetter, srv.URL, "/api/v1/auth/password/reset", "", map[string]string{"email": "intern@example.test"}, 200)
	waitForLetter(t, mailer, 2)
	codeValue := mailer.lastCode(t)

	// Несуществующая почта отвечает так же и молчит.
	post(t, &http.Client{}, srv.URL, "/api/v1/auth/password/reset", "", map[string]string{"email": "nobody@example.test"}, 200)
	if len(mailer.letters()) != 2 {
		t.Fatal("reset for unknown email must not send a letter")
	}

	// Чужое окно не подтверждает сброс.
	if got := post(t, &http.Client{}, srv.URL, "/api/v1/auth/password/confirm", "", map[string]string{"code": codeValue, "newPassword": "NewPassword123!"}, 400); code(got) == "" {
		t.Fatal("foreign window must be rejected with an error code")
	}

	// Владелец кода меняет пароль и сразу входит.
	ok := post(t, resetter, srv.URL, "/api/v1/auth/password/confirm", "", map[string]string{"code": codeValue, "newPassword": "NewPassword123!"}, 200)
	if ok["accessToken"] == "" {
		t.Fatal("no accessToken after password reset")
	}

	// Старые сессии отозваны.
	req(t, h.client, srv.URL, "POST", "/api/v1/auth/refresh", "", nil, 401)

	// Вход по новому паролю работает.
	jar2, _ := cookiejar.New(nil)
	next := &http.Client{Jar: jar2, Timeout: 5 * time.Second}
	post(t, next, srv.URL, "/api/v1/auth/login", "", map[string]string{"email": "intern@example.test", "password": "NewPassword123!"}, 200)
	waitForLetter(t, mailer, 3)
	codeValue = mailer.lastCode(t)
	req(t, next, srv.URL, "POST", "/api/v1/auth/confirm", "", map[string]string{"code": codeValue}, 200)

	// Код логина нельзя потратить на смену пароля.
	if got := post(t, next, srv.URL, "/api/v1/auth/password/confirm", "", map[string]string{"code": codeValue, "newPassword": "OtherPassword1!"}, 400); code(got) == "" {
		t.Fatal("login code must not be accepted as a reset code")
	}
}
