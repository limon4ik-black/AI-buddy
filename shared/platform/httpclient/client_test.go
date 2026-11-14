package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallErrors(t *testing.T) {
	client := &http.Client{}
	if _, err := Call[any](context.Background(), client, "GET", ":bad", "", "", nil); err == nil {
		t.Fatal("invalid URL accepted")
	}
	if _, err := Call[any](context.Background(), client, "GET", "http://localhost", "", "", make(chan int)); err == nil {
		t.Fatal("invalid body accepted")
	}
	invalid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("not json")) }))
	defer invalid.Close()
	if _, err := Call[any](context.Background(), client, "GET", invalid.URL, "", "", nil); err == nil {
		t.Fatal("invalid JSON accepted")
	}
}
