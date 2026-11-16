package features_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/middleware"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/security"
	boardrepo "github.com/limon4ik-black/AI-buddy/note-service/internal/features/board/repo/pg"
	boardservice "github.com/limon4ik-black/AI-buddy/note-service/internal/features/board/service"
	boardhttp "github.com/limon4ik-black/AI-buddy/note-service/internal/features/board/transport/http"
	noterepo "github.com/limon4ik-black/AI-buddy/note-service/internal/features/note/repo/pg"
	noteservice "github.com/limon4ik-black/AI-buddy/note-service/internal/features/note/service"
	notehttp "github.com/limon4ik-black/AI-buddy/note-service/internal/features/note/transport/http"
	"go.uber.org/zap"
)

// The real HTTP -> service -> PostgreSQL path verifies ownership and version
// conflicts together. Every test owns a disposable database, never app volumes.
func TestPersonalWorkspaceHTTP(t *testing.T) {
	dsn := os.Getenv("TEST_DB_ADMIN_URL")
	if dsn == "" {
		t.Fatal("TEST_DB_ADMIN_URL is required; use make test")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := fmt.Sprintf("workspace_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(ctx, "DROP DATABASE "+name+" WITH (FORCE)")
	u, _ := url.Parse(dsn)
	u.Path = "/" + name
	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for _, migration := range []string{"000001_init.up.sql", "000002_boards.up.sql"} {
		sql, err := os.ReadFile(filepath.Join("..", "..", "migrations", migration))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, string(sql)); err != nil {
			t.Fatal(err)
		}
	}
	keys, err := security.ParseKeys(base64.StdEncoding.EncodeToString(make([]byte, ed25519.SeedSize)), "")
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS([]string{" http://workspace.test ", ""}), middleware.Logger(zap.NewNop()), middleware.Recover(zap.NewNop()))
	api := r.Group("/api/v1", middleware.Auth(keys))
	notehttp.New(noteservice.New(noterepo.New(pool, time.Second))).Register(api)
	boardhttp.New(boardservice.New(boardrepo.New(pool, time.Second))).Register(api)
	r.GET("/panic", func(c *gin.Context) { panic("test panic") })
	call := func(method, path, body, token string, status int) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://workspace.test")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != status {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, status, w.Body.String())
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "http://workspace.test" {
			t.Fatal("missing CORS header")
		}
		var result map[string]any
		if w.Body.Len() > 0 && json.Unmarshal(w.Body.Bytes(), &result) != nil {
			t.Fatal("invalid JSON")
		}
		return result
	}
	a, b := keys.Sign("auth-user-a", "a@test"), keys.Sign("auth-user-b", "b@test")
	call("GET", "/panic", "", "", 500)
	call("OPTIONS", "/api/v1/notes", "", "", 204)
	for _, kind := range []string{"notes", "boards"} {
		t.Run(kind, func(t *testing.T) {
			path := "/api/v1/" + kind
			valid := `{"content":"  Личная заметка  "}`
			changed := `{"content":"Изменено","version":1}`
			if kind == "boards" {
				valid = `{"title":"  Моя доска  ","scene":{"elements":[]}}`
				changed = `{"title":"Изменено","scene":{"elements":[{"id":"shape"}]},"version":1}`
			}
			call("GET", path, "", "", 401)
			call("GET", path, "", "invalid", 401)
			if len(call("GET", path, "", a, 200)["items"].([]any)) != 0 {
				t.Fatal("new owner's list is not empty")
			}
			obj := call("POST", path, valid, a, 201)
			id := obj["id"].(string)
			if obj["userId"] != "auth-user-a" || obj["version"] != float64(1) {
				t.Fatal("incorrect owner/version")
			}
			one := path + "/" + id
			call("GET", one, "", a, 200)
			if len(call("GET", path+"?limit=1&offset=0", "", a, 200)["items"].([]any)) != 1 {
				t.Fatal("created item missing from list")
			}
			if len(call("GET", path, "", b, 200)["items"].([]any)) != 0 {
				t.Fatal("another owner sees item")
			}
			for _, method := range []string{"GET", "PATCH", "DELETE"} {
				call(method, one, changed, b, 404)
			}
			updated := call("PATCH", one, changed, a, 200)
			if updated["version"] != float64(2) {
				t.Fatal("version did not increment")
			}
			if call("GET", one, "", a, 200)["version"] != float64(2) {
				t.Fatal("update did not persist")
			}
			if call("PATCH", one, changed, a, 409)["error"].(map[string]any)["code"] != "version_conflict" {
				t.Fatal("missing conflict code")
			}
			for _, q := range []string{"?limit=-1", "?limit=101", "?offset=-1", "?limit=x"} {
				call("GET", path+q, "", a, 400)
			}
			for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
				target := path
				if method == "PATCH" || method == "DELETE" {
					target = one
				}
				body := valid
				if method == "PATCH" {
					body = changed
				}
				call(method, target+"?user_id=other", body, a, 403)
				call(method, target+"?user_id="+strings.Repeat("x", 65), valid, a, 400)
			}
			for _, method := range []string{"GET", "PATCH", "DELETE"} {
				call(method, path+"/"+strings.Repeat("x", 65), changed, a, 400)
			}
			call("PATCH", one, `{}`, a, 400)
			for _, method := range []string{"POST", "PATCH"} {
				target := path
				if method == "PATCH" {
					target = one
				}
				for _, body := range []string{`{`, `{"unknown":1}`, valid + ` {}`} {
					call(method, target, body, a, 400)
				}
			}
			if kind == "notes" {
				for _, text := range []string{" ", strings.Repeat("я", 20001)} {
					body, _ := json.Marshal(map[string]any{"content": text, "version": 2})
					call("POST", path, string(body), a, 400)
					call("PATCH", one, string(body), a, 400)
				}
			} else {
				for _, scene := range []string{`null`, `[]`, `"string"`} {
					body := `{"title":"test","scene":` + scene + `,"version":2}`
					call("POST", path, body, a, 400)
					call("PATCH", one, body, a, 400)
				}
				body, _ := json.Marshal(map[string]any{"title": strings.Repeat("я", 201), "scene": map[string]any{}, "version": 2})
				call("POST", path, string(body), a, 400)
				call("PATCH", one, string(body), a, 400)
			}
			call("DELETE", one, "", a, 204)
			for _, method := range []string{"GET", "PATCH", "DELETE"} {
				call(method, one, changed, a, 404)
			}
			if len(call("GET", path, "", a, 200)["items"].([]any)) != 0 {
				t.Fatal("deleted item remains")
			}
		})
	}
	// Database failures must remain errors, never empty successful lists.
	pool.Close()
	for _, kind := range []string{"notes", "boards"} {
		path := "/api/v1/" + kind
		body := `{"content":"test","version":1}`
		if kind == "boards" {
			body = `{"title":"test","scene":{},"version":1}`
		}
		for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
			target := path
			requestBody := body
			if method == "POST" {
				if kind == "notes" { requestBody = `{"content":"test"}` } else { requestBody = `{"title":"test","scene":{}}` }
			}
			if method == "PATCH" || method == "DELETE" {
				target += "/missing"
			}
			call(method, target, requestBody, a, 500)
		}
	}
}
