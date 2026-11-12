package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
)

func TestContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("request contract mismatch")
		}
		var b struct {
			Message string      `json:"message"`
			History []chat.Turn `json:"history"`
		}
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			t.Error(err)
		}
		if b.Message != "new" || len(b.History) != 1 || b.History[0].Content != "old" {
			t.Error(b)
		}
		_, _ = w.Write([]byte(`{"answer":"reply","sources":[{"id":"1"}]}`))
	}))
	defer server.Close()
	c := Client{URL: server.URL, Token: "test-token", HTTP: server.Client()}
	answer, err := c.Answer(context.Background(), "new", []chat.Turn{{Role: "user", Content: "old"}})
	if err != nil || answer != "reply" {
		t.Fatal(answer, err)
	}
}
func TestFailures(t *testing.T) {
	for _, body := range []string{`{}`, `{"answer":null}`, `{"answer":" "}`, `{"answer":3}`, `broken`, `{"answer":"ok"} {}`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer server.Close()
			c := Client{URL: server.URL, HTTP: server.Client()}
			if _, err := c.Answer(context.Background(), "hi", nil); err == nil {
				t.Fatal("expected invalid response")
			}
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(502); _, _ = w.Write([]byte("secret")) }))
	defer server.Close()
	c := Client{URL: server.URL, HTTP: server.Client()}
	if _, err := c.Answer(context.Background(), "hi", nil); err == nil || err.Error() == "secret" {
		t.Fatal(err)
	}
}
func TestTimeoutAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer server.Close()
	c := Client{URL: server.URL, HTTP: &http.Client{Timeout: 20 * time.Millisecond}}
	if _, err := c.Answer(context.Background(), "hi", nil); err == nil {
		t.Fatal("expected timeout")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Answer(ctx, "hi", nil); err == nil {
		t.Fatal("expected cancellation")
	}
}
