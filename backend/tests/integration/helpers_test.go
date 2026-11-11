package integration

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/limon4ik-black/AI-buddy/backend/internal/auth"

	chatsvc "github.com/limon4ik-black/AI-buddy/backend/internal/chat"
	"github.com/limon4ik-black/AI-buddy/backend/internal/corpmock"
	"github.com/limon4ik-black/AI-buddy/backend/internal/gateway"
	"github.com/limon4ik-black/AI-buddy/backend/internal/info"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
	"github.com/limon4ik-black/AI-buddy/backend/internal/ragmock"
	"github.com/limon4ik-black/AI-buddy/backend/migrations"
)

var contractTargets sync.Map

var contractFixtures []map[string]any

func TestMain(m *testing.M) {
	code := m.Run()
	if output := os.Getenv("CONTRACT_FIXTURES_OUT"); output != "" && code == 0 {
		b, _ := json.MarshalIndent(contractFixtures, "", "  ")
		if e := os.WriteFile(output, b, 0600); e != nil {
			fmt.Fprintln(os.Stderr, e)
			code = 1
		}
	}
	os.Exit(code)
}

type harness struct {
	auth                              *auth.Service
	info                              *info.Service
	chat                              *chatsvc.Service
	keys                              *security.Keys
	gateway, infoURL, ragURL, corpURL string
	client                            *http.Client
	token                             string
}

func database(t *testing.T, name string) *pgxpool.Pool {
	t.Helper()
	adminURL := os.Getenv("TEST_DB_ADMIN_URL")
	if adminURL == "" {
		t.Fatal("Use bash scripts/test-backend.sh (TEST_DB_ADMIN_URL is required)")
	}
	admin, e := pgxpool.New(context.Background(), adminURL)
	if e != nil {
		t.Fatal(e)
	}
	dbName := "buddy_test_" + name + "_" + idgen.New()
	if _, e = admin.Exec(context.Background(), "CREATE DATABASE "+dbName); e != nil {
		t.Fatal(e)
	}
	u, _ := url.Parse(adminURL)
	u.Path = "/" + dbName
	params := u.Query()
	params.Set("pool_max_conns", "2")
	u.RawQuery = params.Encode()
	updates := [][]byte{}
	if name == "info" {
		updates = append(updates, migrations.InfoMVP)
	}
	db, e := postgres.Open(context.Background(), u.String(), name, map[string][]byte{"auth": migrations.Auth, "info": migrations.Info, "chat": migrations.Chat}[name], updates...)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		db.Close()
		_, _ = admin.Exec(context.Background(), "DROP DATABASE "+dbName+" WITH (FORCE)")
		admin.Close()
	})
	return db
}

func setup(t *testing.T) *harness {
	t.Helper()
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	k := &security.Keys{Private: priv, Public: pub, Now: time.Now}
	h := &harness{keys: k, client: &http.Client{Timeout: 3 * time.Second}}
	serve := func(handler http.Handler) string { s := httptest.NewServer(handler); t.Cleanup(s.Close); return s.URL }
	h.ragURL = serve(ragmock.New("rag-secret"))
	h.corpURL = serve(corpmock.New("corp-secret"))
	h.auth = &auth.Service{DB: database(t, "auth"), Keys: k}
	if e := h.auth.SeedDemo(context.Background()); e != nil {
		t.Fatal(e)
	}
	authURL := serve(h.auth.Handler())
	h.info = &info.Service{DB: database(t, "info"), Keys: k, GatewayToken: "gateway-secret", ChatToken: "chat-secret", IndexerToken: "index-secret", CorpToken: "corp-secret", CorpURL: h.corpURL, Client: h.client}
	h.infoURL = serve(h.info.Handler())
	h.chat = &chatsvc.Service{DB: database(t, "chat"), Keys: k, GatewayToken: "gateway-secret", ChatToken: "chat-secret", RAGToken: "rag-secret", InfoURL: h.infoURL, RAGURL: h.ragURL, Client: h.client}
	chatURL := serve(h.chat.Handler())
	gateway, e := gateway.New(gateway.Config{Keys: k, Token: "gateway-secret", Origin: "http://localhost:5173", AuthURL: authURL, InfoURL: h.infoURL, ChatURL: chatURL})
	if e != nil {
		t.Fatal(e)
	}
	h.gateway = serve(gateway)
	for target, spec := range map[string]string{h.gateway: "public", h.infoURL: "info-internal", h.ragURL: "rag-internal", h.corpURL: "corporate-internal"} {
		contractTargets.Store(target, spec)
	}
	jar, _ := cookiejar.New(nil)
	h.client = &http.Client{Jar: jar, Timeout: 5 * time.Second}
	h.token = k.Sign("demo-intern", "intern@example.test")
	return h
}

func req(t *testing.T, client *http.Client, base, method, path, token string, body any, status int, headers ...string) []byte {
	t.Helper()
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	r, e := http.NewRequest(method, base+path, bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	for j := 0; j < len(headers); j += 2 {
		r.Header.Set(headers[j], headers[j+1])
	}
	res, e := client.Do(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode != status {
		t.Fatalf("%s %s: got %d want %d: %s", method, path, res.StatusCode, status, data)
	}
	if res.Header.Get("X-Request-ID") == "" && strings.Contains(path, "/api/") {
		t.Fatal("Missing request id")
	}
	if spec, ok := contractTargets.Load(base); ok && len(data) > 0 && res.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		var payload any
		if json.Unmarshal(data, &payload) == nil {
			if m, ok := payload.(map[string]any); ok {
				if _, present := m["accessToken"]; present {
					m["accessToken"] = "[redacted test token]"
				}
			}
			contractFixtures = append(contractFixtures, map[string]any{"spec": spec, "path": strings.Split(path, "?")[0], "method": strings.ToLower(method), "status": res.StatusCode, "body": payload})
		}
	}
	return data
}

func unpack[T any](t *testing.T, b []byte) T {
	t.Helper()
	var v T
	if e := json.Unmarshal(b, &v); e != nil {
		t.Fatalf("%s: %v", b, e)
	}
	return v
}

func obj(args ...any) map[string]any {
	m := map[string]any{}
	for j := 0; j < len(args); j += 2 {
		m[args[j].(string)] = args[j+1]
	}
	return m
}
