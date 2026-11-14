package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/testutil"
)

func TestREST(t *testing.T) {
	repo := testutil.New()
	h := (&Handler{Service: &chat.Service{Repository: repo}, Timeout: time.Second}).Routes()
	u := uuid.New().String()
	other := uuid.New().String()
	request := func(method, path, user string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, nil)
		if user != "" {
			r.Header.Set("X-User-ID", user)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, v := range []string{"", " bad "} {
		if w := request("GET", "/chats", v); w.Code != 401 {
			t.Fatal(w)
		}
	}
	if w := request("GET", "/chats", "intern-egor"); w.Code != 200 {
		t.Fatal("signed string subject must be accepted", w)
	}
	if w := request("GET", "/chats", u); w.Code != 200 || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatal(w)
	}
	w := request("POST", "/chats", u)
	if w.Code != 201 {
		t.Fatal(w)
	}
	var created chat.Created
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Title != "Новый чат" || created.ID == uuid.Nil || created.CreatedAt.IsZero() {
		t.Fatal(created)
	}
	path := "/chats/" + created.ID.String()
	for _, test := range []struct {
		method, path, user string
		status             int
	}{
		{"GET", path + "/messages", u, 200}, {"GET", path + "/messages", other, 404}, {"DELETE", path, other, 404},
		{"GET", "/chats/not-uuid/messages", u, 400}, {"DELETE", "/chats/not-uuid", u, 400},
		{"GET", "/chats/" + uuid.NewString() + "/messages", u, 404}, {"DELETE", path, u, 204}, {"GET", path + "/messages", u, 404},
	} {
		if w := request(test.method, test.path, test.user); w.Code != test.status {
			t.Fatalf("%+v: %d %s", test, w.Code, w.Body.String())
		}
	}
	repo.Failure = errors.New("postgres://secret")
	w = request("GET", "/chats", u)
	if w.Code != 503 || strings.Contains(w.Body.String(), "secret") {
		t.Fatal(w)
	}
}
func TestWebSocketMultipleChatsAndErrors(t *testing.T) {
	repo := testutil.New()
	u := uuid.New()
	one, _ := repo.Create(context.Background(), u)
	two, _ := repo.Create(context.Background(), u)
	foreign, _ := repo.Create(context.Background(), uuid.New())
	svc := &chat.Service{Repository: repo, AI: testutil.AIFunc(func(_ context.Context, s string, _ []chat.Turn) (string, error) {
		if s == "fail" {
			return "", errors.New("secret")
		}
		return "reply " + s, nil
	})}
	server := httptest.NewServer((&Handler{Service: svc, Timeout: time.Second}).Routes())
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, server.URL+"/ws", &websocket.DialOptions{HTTPHeader: http.Header{"X-User-Id": []string{u.String()}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	for _, test := range []struct{ body, kind, content string }{
		{`{"chat_id":"` + one.ID.String() + `","content":"first"}`, "message", "reply first"},
		{`{"chat_id":"` + two.ID.String() + `","content":"second"}`, "message", "reply second"},
		{`{"chat_id":"` + foreign.ID.String() + `","content":"blocked"}`, "error", ""},
		{`{"chat_id":"bad","content":"hello"}`, "error", ""},
		{`{"chat_id":"` + one.ID.String() + `","content":" "}`, "error", ""},
		{`{"chat_id":"` + one.ID.String() + `","content":"fail"}`, "error", ""},
		{`not json`, "error", ""},
	} {
		if err = conn.Write(ctx, websocket.MessageText, []byte(test.body)); err != nil {
			t.Fatal(err)
		}
		_, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var e event
		if err = json.Unmarshal(data, &e); err != nil {
			t.Fatal(err)
		}
		if e.Type != test.kind || e.Content != test.content || strings.Contains(string(data), "secret") {
			t.Fatalf("%s: %s", test.body, data)
		}
	}
	history, _ := repo.History(ctx, u, one.ID)
	if len(history) != 3 || history[2].Content != "fail" {
		t.Fatal(history)
	}
	history, _ = repo.History(ctx, u, two.ID)
	if len(history) != 2 {
		t.Fatal(history)
	}
}
func TestWebSocketDisconnectCancelsAI(t *testing.T) {
	repo := testutil.New()
	u := uuid.New()
	c, _ := repo.Create(context.Background(), u)
	started := make(chan struct{})
	canceled := make(chan struct{})
	svc := &chat.Service{Repository: repo, AI: testutil.AIFunc(func(ctx context.Context, _ string, _ []chat.Turn) (string, error) {
		close(started)
		<-ctx.Done()
		close(canceled)
		return "", ctx.Err()
	})}
	handler := &Handler{Service: svc, Timeout: time.Second * 10}
	server := httptest.NewServer(handler.Routes())
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, server.URL+"/ws", &websocket.DialOptions{HTTPHeader: http.Header{"X-User-Id": []string{u.String()}}})
	if err != nil {
		t.Fatal(err)
	}
	if err = conn.Write(ctx, websocket.MessageText, []byte(`{"chat_id":"`+c.ID.String()+`","content":"hello"}`)); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("AI not started")
	}
	_ = conn.CloseNow()
	select {
	case <-canceled:
	case <-ctx.Done():
		t.Fatal("AI not canceled on disconnect")
	}
	handler.Wait()
}

func TestWebSocketProcessesMessagesSequentially(t *testing.T) {
	repo := testutil.New()
	u := uuid.New()
	c, _ := repo.Create(context.Background(), u)
	started := make(chan string, 2)
	release := make(chan struct{})
	svc := &chat.Service{Repository: repo, AI: testutil.AIFunc(func(ctx context.Context, content string, history []chat.Turn) (string, error) {
		started <- content
		if content == "first" {
			select {
			case <-release:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		if content == "second" && (len(history) != 2 || history[1].Role != "assistant") {
			t.Error("second message did not see the first completed exchange")
		}
		return "reply " + content, nil
	})}
	server := httptest.NewServer((&Handler{Service: svc, Timeout: time.Second}).Routes())
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, server.URL+"/ws", &websocket.DialOptions{HTTPHeader: http.Header{"X-User-Id": []string{u.String()}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	for _, content := range []string{"first", "second"} {
		data, _ := json.Marshal(map[string]string{"chat_id": c.ID.String(), "content": content})
		if err = conn.Write(ctx, websocket.MessageText, data); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case content := <-started:
		if content != "first" {
			t.Fatal(content)
		}
	case <-ctx.Done():
		t.Fatal("first not started")
	}
	select {
	case content := <-started:
		t.Fatalf("concurrent AI call: %s", content)
	default:
	}
	close(release)
	for _, content := range []string{"first", "second"} {
		_, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var response event
		if err = json.Unmarshal(data, &response); err != nil {
			t.Fatal(err)
		}
		if response.Type != "message" || response.Content != "reply "+content {
			t.Fatal(string(data))
		}
	}
}
