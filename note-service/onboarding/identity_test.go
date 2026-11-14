package onboarding

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/shared/platform/security"
)

func identityDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DB_ADMIN_URL")
	if dsn == "" {
		t.Skip("TEST_DB_ADMIN_URL is required for PostgreSQL identity tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	name := "note_identity_" + idgen.New()
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	db, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close(); _, _ = admin.Exec(ctx, "DROP DATABASE "+name+" WITH (FORCE)"); admin.Close() })
	return db
}

func identityKeys(t *testing.T) *security.Keys {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return &security.Keys{Private: priv, Public: pub, Now: time.Now}
}

func identityRequest(t *testing.T, client *http.Client, method, endpoint, jwt, serviceToken string, body any, status int) []byte {
	t.Helper()
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	r, err := http.NewRequest(method, endpoint, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if jwt != "" {
		r.Header.Set("Authorization", "Bearer "+jwt)
	}
	if serviceToken != "" {
		r.Header.Set("X-Service-Token", serviceToken)
	}
	r.Header.Set("Content-Type", "application/json")
	// Client-selected identity headers never override the signed subject.
	r.Header.Set("X-User-ID", "attacker-selected-id")
	response, err := client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err = io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, endpoint, response.StatusCode, status, data)
	}
	return data
}

func identityFixture(t *testing.T, db *pgxpool.Pool, schema string) {
	t.Helper()
	_, err := db.Exec(context.Background(), `INSERT INTO `+schema+`.knowledge_documents(id,team_id,title,content,category,visibility,source_url) VALUES
 ('common-doc',NULL,'Welcome','Common guide','onboarding','company','https://example.test/common'),
 ('backend-doc','backend','Backend','Backend guide','onboarding','team','https://example.test/backend'),
 ('analytics-doc','analytics','Analytics','Private analytics guide','onboarding','team','https://example.test/analytics');
 INSERT INTO `+schema+`.checklist_tasks(id,team_id,title,description,period,sort_order,document_id,source_url) VALUES
 ('common-task',NULL,'Common step','Description','week1',1,'common-doc','https://example.test/common'),
 ('backend-task','backend','Backend step','Description','week1',2,'backend-doc','https://example.test/backend'),
 ('analytics-task','analytics','Analytics step','Description','week1',3,'analytics-doc','https://example.test/analytics');
 INSERT INTO `+schema+`.user_task_progress(user_id,task_id,status,completed_at) VALUES('intern-egor','common-task','completed',now())`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestContextRejectsForeignAuthIDBeforeEmployeeOrDatabase(t *testing.T) {
	keys := identityKeys(t)
	var calls atomic.Int32
	employee := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(500) }))
	defer employee.Close()
	s := &Service{Keys: keys, ChatToken: "chat-token", IndexerToken: "index-token", EmployeeURL: employee.URL, Client: employee.Client()}
	server := httptest.NewServer(s.Handler())
	defer server.Close()
	jwt := keys.Sign("auth-user", "user@example.test")
	identityRequest(t, server.Client(), "GET", server.URL+"/internal/v1/users/other/context", jwt, "chat-token", nil, 403)
	identityRequest(t, server.Client(), "GET", server.URL+"/internal/v1/users/auth-user/context", jwt, "index-token", nil, 401)
	identityRequest(t, server.Client(), "GET", server.URL+"/internal/v1/users/auth-user/context", "", "chat-token", nil, 401)
	identityRequest(t, server.Client(), "GET", server.URL+"/internal/v1/users/auth-user/context", "invalid-jwt", "chat-token", nil, 401)
	if calls.Load() != 0 {
		t.Fatal("unauthorized identity reached Employee")
	}
}

func TestCanonicalAuthIdentityPreservesProgressAndTeamACL(t *testing.T) {
	db := identityDB(t)
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	identityFixture(t, db, "onboarding")
	keys := identityKeys(t)
	authID := "registered-" + idgen.New()
	jwt := keys.Sign(authID, "backend@buddy.test")
	var calls atomic.Int32
	employee := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/internal/v1/users/") {
			calls.Add(1)
			claims, err := keys.Verify(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			if err != nil || claims.Subject != authID || r.Header.Get("Authorization") != "Bearer "+jwt || r.Header.Get("X-Service-Token") != "chat-token" || r.URL.Path != "/internal/v1/users/"+authID+"/context" {
				t.Error("Employee received altered/untrusted identity")
				w.WriteHeader(403)
				return
			}
			_ = json.NewEncoder(w).Encode(contracts.UserContext{Profile: contracts.Profile{UserID: authID, FullName: "Егор", TeamID: "backend"}, Team: contracts.TeamDetail{Team: contracts.Team{ID: "backend"}}, Role: "intern", Tools: []contracts.Tool{}})
			return
		}
		if r.Header.Get("X-Service-Token") != "index-token" {
			w.WriteHeader(401)
			return
		}
		if r.URL.Path == "/internal/v1/teams" {
			_, _ = w.Write([]byte(`{"items":[{"id":"backend"},{"id":"analytics"}]}`))
			return
		}
		if r.URL.Path == "/internal/v1/services" {
			_, _ = w.Write([]byte(`{"items":[]}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer employee.Close()
	s := &Service{DB: db, Keys: keys, GatewayToken: "gateway-token", ChatToken: "chat-token", IndexerToken: "index-token", EmployeeURL: employee.URL, Client: employee.Client()}
	server := httptest.NewServer(s.Handler())
	defer server.Close()
	endpoint := server.URL + "/internal/v1/users/" + authID + "/context"
	var got contracts.UserContext
	if err := json.Unmarshal(identityRequest(t, server.Client(), "GET", endpoint, jwt, "chat-token", nil, 200), &got); err != nil {
		t.Fatal(err)
	}
	if got.Profile.UserID != authID || got.Profile.TeamID != "backend" || got.Onboarding.Total != 2 || got.Onboarding.Completed != 1 {
		t.Fatalf("identity/profile/progress changed: %+v", got)
	}
	if slices.Contains(got.AllowedDocumentIDs, "analytics-doc") || !slices.Contains(got.AllowedDocumentIDs, "backend-doc") {
		t.Fatal("document ACL not based on Employee team", got.AllowedDocumentIDs)
	}
	for _, task := range got.Onboarding.Tasks {
		if task.ID == "analytics-task" {
			t.Fatal("foreign team checklist leaked")
		}
	}
	identityRequest(t, server.Client(), "PATCH", server.URL+"/api/v1/checklist/tasks/analytics-task", jwt, "gateway-token", map[string]string{"status": "completed"}, 404)
	identityRequest(t, server.Client(), "PATCH", server.URL+"/api/v1/checklist/tasks/common-task", jwt, "gateway-token", map[string]string{"status": "pending"}, 200)
	identityRequest(t, server.Client(), "PATCH", server.URL+"/api/v1/checklist/tasks/backend-task", jwt, "gateway-token", map[string]string{"status": "completed", "teamId": "analytics"}, 400)
	var checklist contracts.Checklist
	if err := json.Unmarshal(identityRequest(t, server.Client(), "GET", server.URL+"/api/v1/checklist?teamId=analytics", jwt, "gateway-token", nil, 200), &checklist); err != nil {
		t.Fatal(err)
	}
	if checklist.Total != 2 || checklist.Completed != 0 {
		t.Fatal("client team override or replayed template progress", checklist)
	}
	identityRequest(t, server.Client(), "GET", endpoint, jwt, "chat-token", nil, 200)
	var links int
	if err := db.QueryRow(context.Background(), "SELECT count(*) FROM onboarding.identity_links WHERE user_id=$1 AND source_user_id='intern-egor'", authID).Scan(&links); err != nil || links != 1 {
		t.Fatal("identity mapping not persistent", links, err)
	}
	var status string
	if err := db.QueryRow(context.Background(), "SELECT status FROM onboarding.user_task_progress WHERE user_id=$1 AND task_id='common-task'", authID).Scan(&status); err != nil || status != "pending" {
		t.Fatal("edited progress overwritten", status, err)
	}
	if calls.Load() < 3 {
		t.Fatal("Employee identity was not checked on requests")
	}
}

func TestEmployeeIdentityMismatchDoesNotCreateNoteLink(t *testing.T) {
	db := identityDB(t)
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	keys := identityKeys(t)
	employee := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"profile":{"userId":"someone-else","teamId":"analytics"}}`))
	}))
	defer employee.Close()
	s := &Service{DB: db, Keys: keys, ChatToken: "chat-token", EmployeeURL: employee.URL, Client: employee.Client()}
	server := httptest.NewServer(s.Handler())
	defer server.Close()
	identityRequest(t, server.Client(), "GET", server.URL+"/internal/v1/users/auth-user/context", keys.Sign("auth-user", "user@example.test"), "chat-token", nil, 502)
	var count int
	if err := db.QueryRow(context.Background(), "SELECT count(*) FROM onboarding.identity_links").Scan(&count); err != nil || count != 0 {
		t.Fatal("foreign identity wrote note state", count, err)
	}
}

func TestLegacyImportPreservesCanonicalTextIDsAndDoesNotReplay(t *testing.T) {
	source, destination := identityDB(t), identityDB(t)
	ctx := context.Background()
	if err := Migrate(ctx, source); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, destination); err != nil {
		t.Fatal(err)
	}
	// ImportLegacy reads the former Info public schema. This source fixture has
	// exactly those row shapes, including nullable team/document timestamps.
	if _, err := source.Exec(ctx, `CREATE TABLE public.knowledge_documents (LIKE onboarding.knowledge_documents INCLUDING ALL); CREATE TABLE public.checklist_tasks (LIKE onboarding.checklist_tasks INCLUDING ALL); CREATE TABLE public.user_task_progress (LIKE onboarding.user_task_progress INCLUDING ALL)`); err != nil {
		t.Fatal(err)
	}
	identityFixture(t, source, "public")
	canonical := "auth-not-a-uuid-" + idgen.New()
	if _, err := source.Exec(ctx, "INSERT INTO public.user_task_progress(user_id,task_id,status,completed_at) VALUES($1,'backend-task','completed',now())", canonical); err != nil {
		t.Fatal(err)
	}
	if err := ImportLegacy(ctx, source, destination); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := destination.QueryRow(ctx, "SELECT status FROM onboarding.user_task_progress WHERE user_id=$1 AND task_id='backend-task'", canonical).Scan(&status); err != nil || status != "completed" {
		t.Fatal("Auth string ID lost in import", status, err)
	}
	if _, err := destination.Exec(ctx, "UPDATE onboarding.user_task_progress SET status='pending',completed_at=NULL WHERE user_id=$1 AND task_id='backend-task'", canonical); err != nil {
		t.Fatal(err)
	}
	if _, err := destination.Exec(ctx, "UPDATE onboarding.knowledge_documents SET title='User changed title' WHERE id='common-doc'"); err != nil {
		t.Fatal(err)
	}
	if err := ImportLegacy(ctx, source, destination); err != nil {
		t.Fatal(err)
	}
	if err := destination.QueryRow(ctx, "SELECT status FROM onboarding.user_task_progress WHERE user_id=$1 AND task_id='backend-task'", canonical).Scan(&status); err != nil || status != "pending" {
		t.Fatal("import replay overwrote personal progress", status, err)
	}
	var title string
	if err := destination.QueryRow(ctx, "SELECT title FROM onboarding.knowledge_documents WHERE id='common-doc'").Scan(&title); err != nil || title != "User changed title" {
		t.Fatal("import replay overwrote Note data", title, err)
	}
}
