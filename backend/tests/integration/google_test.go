package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func fakeGoogle(t *testing.T, sub, email string) (*oauth2.Config, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		http.Redirect(w, r, q.Get("redirect_uri")+"?code=good-code&state="+q.Get("state"), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"fake-access","token_type":"Bearer"}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sub": sub, "email": email, "name": "Google Tester", "email_verified": true})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &oauth2.Config{
		ClientID: "test-client", ClientSecret: "test-secret",
		Endpoint: oauth2.Endpoint{AuthURL: srv.URL + "/authorize", TokenURL: srv.URL + "/token"},
	}, srv.URL
}

func googleLogin(t *testing.T, client *http.Client, base string) *http.Response {
	t.Helper()
	nr := &http.Client{
		Jar:     client.Jar,
		Timeout: client.Timeout,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if req.URL.Path == "/" && req.URL.RawQuery == "" {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := nr.Get(base + "/api/v1/auth/google/login")
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func refreshAccessToken(t *testing.T, client *http.Client, base string) string {
	t.Helper()
	raw := req(t, client, base, "POST", "/api/v1/auth/refresh", "", nil, 200)
	var out struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.AccessToken == "" {
		t.Fatalf("refresh: %s", raw)
	}
	return out.AccessToken
}

func meUserID(t *testing.T, client *http.Client, base, access string) string {
	t.Helper()
	raw := req(t, client, base, "GET", "/api/v1/auth/me", access, nil, 200)
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out.ID
}

func TestGoogleAuthCreatesAccounts(t *testing.T) {
	h := setup(t)
	conf, fakeURL := fakeGoogle(t, "google-sub-1", "google@example.test")
	server := httptest.NewServer(h.auth.Handler())
	t.Cleanup(server.Close)
	conf.RedirectURL = server.URL + "/api/v1/auth/google/callback"
	h.auth.Google = conf
	h.auth.GoogleUserInfo = fakeURL + "/userinfo"

	first := googleLogin(t, h.client, server.URL)
	defer first.Body.Close()
	if first.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(first.Body)
		t.Fatalf("google login status: %d body: %s", first.StatusCode, body)
	}
	if loc := first.Header.Get("Location"); !strings.HasSuffix(loc, "/") {
		t.Fatalf("expected redirect to frontend root, got %q", loc)
	}
	access := refreshAccessToken(t, h.client, server.URL)
	userID := meUserID(t, h.client, server.URL, access)
	me := req(t, h.client, server.URL, "GET", "/api/v1/auth/me", access, nil, 200)
	var meOut struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(me, &meOut); err != nil {
		t.Fatal(err)
	}
	if meOut.Email != "google@example.test" {
		t.Fatalf("/me mismatch: body %s", me)
	}

	again := googleLogin(t, h.client, server.URL)
	again.Body.Close()
	if again.StatusCode != http.StatusFound {
		t.Fatalf("second login status: %d", again.StatusCode)
	}
	if againID := meUserID(t, h.client, server.URL, refreshAccessToken(t, h.client, server.URL)); againID != userID {
		t.Fatalf("same sub produced different users: %s vs %s", againID, userID)
	}

	conf2, fakeURL2 := fakeGoogle(t, "google-sub-2", "google@example.test")
	conf2.RedirectURL = server.URL + "/api/v1/auth/google/callback"
	h.auth.Google = conf2
	h.auth.GoogleUserInfo = fakeURL2 + "/userinfo"
	linked := googleLogin(t, h.client, server.URL)
	linked.Body.Close()
	if linked.StatusCode != http.StatusFound {
		t.Fatalf("link login status: %d", linked.StatusCode)
	}
	if linkedID := meUserID(t, h.client, server.URL, refreshAccessToken(t, h.client, server.URL)); linkedID != userID {
		t.Fatalf("email link failed: %s vs %s", linkedID, userID)
	}
}
