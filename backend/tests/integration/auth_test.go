package integration

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

func TestAuthLifecycle(t *testing.T) {
	h := setup(t)
	req(t, h.client, h.gateway, "POST", "/api/v1/auth/login", "", obj("email", "intern@example.test", "password", "wrong-password"), 401)
	req(t, h.client, h.gateway, "POST", "/api/v1/auth/login", "", obj("email", "absent@example.test", "password", "wrong-password"), 401)
	for _, data := range []any{obj("email", "bad", "password", "long-password"), obj("email", "a@example.test", "password", "short"), obj("email", "Name <a@example.test>", "password", "long-password"), obj("extra", true)} {
		req(t, h.client, h.gateway, "POST", "/api/v1/auth/register", "", data, 400)
	}
	// Вход подтверждается ссылкой из письма: регистрация сессию не открывает.
	mailer := &captureMailer{}
	h.auth.Mailer = mailer
	data := obj("email", "NEW@example.test", "password", "LongPassword123!")
	req(t, h.client, h.gateway, "POST", "/api/v1/auth/register", "", data, 201)
	req(t, h.client, h.gateway, "POST", "/api/v1/auth/register", "", data, 409)
	link := mailer.lastToken(t)
	confirmResp, e := noRedirectAtRoot(h.client).Get(h.gateway + "/api/v1/auth/confirm?token=" + link)
	if e != nil {
		t.Fatal(e)
	}
	confirmResp.Body.Close()
	if confirmResp.StatusCode != http.StatusFound {
		t.Fatalf("confirm status: %d", confirmResp.StatusCode)
	}
	raw := req(t, h.client, h.gateway, "POST", "/api/v1/auth/refresh", "", nil, 200)
	token := unpack[struct{ AccessToken string }](t, raw).AccessToken
	req(t, h.client, h.gateway, "GET", "/api/v1/auth/me", token, nil, 200)
	u, _ := url.Parse(h.gateway + "/api/v1/auth/refresh")
	oldCookie := h.client.Jar.Cookies(u)
	raw = req(t, h.client, h.gateway, "POST", "/api/v1/auth/refresh", "", nil, 200)
	token = unpack[struct{ AccessToken string }](t, raw).AccessToken
	plain := &http.Client{}
	req(t, plain, h.gateway, "POST", "/api/v1/auth/refresh", "", nil, 401)
	req(t, plain, h.gateway, "POST", "/api/v1/auth/refresh", "", nil, 401, "Cookie", oldCookie[0].Name+"="+oldCookie[0].Value)
	req(t, h.client, h.gateway, "GET", "/api/v1/auth/me", token, nil, 200)
	req(t, h.client, h.gateway, "POST", "/api/v1/auth/logout", "", nil, 204)
	req(t, h.client, h.gateway, "POST", "/api/v1/auth/logout", "", nil, 204)
	req(t, h.client, h.gateway, "POST", "/api/v1/auth/refresh", "", nil, 401)
	// Повторный вход: снова ссылка, снова подтверждение, снова сессия.
	req(t, h.client, h.gateway, "POST", "/api/v1/auth/login", "", data, 200)
	link = mailer.lastToken(t)
	confirmResp, e = noRedirectAtRoot(h.client).Get(h.gateway + "/api/v1/auth/confirm?token=" + link)
	if e != nil {
		t.Fatal(e)
	}
	confirmResp.Body.Close()
	if confirmResp.StatusCode != http.StatusFound {
		t.Fatalf("re-login confirm status: %d", confirmResp.StatusCode)
	}
	// Exactly one caller can consume a refresh token, even with concurrent delivery.
	cookies := h.client.Jar.Cookies(u)
	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, _ := http.NewRequest("POST", u.String(), nil)
			r.AddCookie(cookies[0])
			res, e := plain.Do(r)
			if e != nil {
				statuses <- 0
				return
			}
			defer res.Body.Close()
			statuses <- res.StatusCode
		}()
	}
	wg.Wait()
	close(statuses)
	sum := 0
	for s := range statuses {
		sum += s
	}
	if sum != 601 {
		t.Fatal("Refresh replay accepted", sum)
	}
	if e := h.auth.SeedDemo(context.Background()); e != nil {
		t.Fatal(e)
	}
}
