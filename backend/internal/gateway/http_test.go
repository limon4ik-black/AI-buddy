package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

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
	return data
}
