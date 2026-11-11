package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

// fakeGoogle mimics the Google OAuth endpoints (authorize, token, userinfo).
// Returns the OAuth config and the fake server base URL.
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

// googleLogin performs the full OAuth round trip through the local fake Google.
func googleLogin(t *testing.T, client *http.Client, base string) map[string]any {
	t.Helper()
	resp, err := client.Get(base + "/api/v1/auth/google/login")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("google login status: %d url: %s body: %s", resp.StatusCode, resp.Request.URL, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("google login body: %s", body)
	}
	return out
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
	user := first["user"].(map[string]any)
	if user["email"] != "google@example.test" || user["role"] != "intern" {
		t.Fatalf("unexpected user: %v", user)
	}
	access, ok := first["accessToken"].(string)
	if !ok || access == "" {
		t.Fatalf("no access token: %v", first)
	}

	// Тот же Google-аккаунт повторно: тот же user_id.
	again := googleLogin(t, h.client, server.URL)
	if again["user"].(map[string]any)["id"] != user["id"] {
		t.Fatalf("same sub produced different users: %v vs %v", again["user"], user)
	}

	// Тот же email с другим Google-аккаунтом: привязка к существующему юзеру.
	conf2, fakeURL2 := fakeGoogle(t, "google-sub-2", "google@example.test")
	conf2.RedirectURL = server.URL + "/api/v1/auth/google/callback"
	h.auth.Google = conf2
	h.auth.GoogleUserInfo = fakeURL2 + "/userinfo"
	linked := googleLogin(t, h.client, server.URL)
	if linked["user"].(map[string]any)["id"] != user["id"] {
		t.Fatalf("email link failed: %v", linked["user"])
	}

	// Выданный access-токен работает на /me.
	me := req(t, h.client, server.URL, "GET", "/api/v1/auth/me", access, nil, 200)
	var meOut struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(me, &meOut); err != nil {
		t.Fatal(err)
	}
	if meOut.ID != user["id"] || meOut.Email != "google@example.test" {
		t.Fatalf("/me mismatch: body %s", me)
	}
}
