package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransportErrorsAndLimits(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("POST /api/test", Handle(func(w http.ResponseWriter, r *http.Request) error {
		_, e := Decode[map[string]string](r)
		if e != nil {
			return e
		}
		JSON(w, 201, map[string]bool{"ok": true})
		return nil
	}))
	s := httptest.NewServer(Middleware(m))
	defer s.Close()
	for _, data := range []string{`{} {}`, `{bad}`, `{"a":"` + strings.Repeat("x", 1<<20) + `"}`} {
		r, _ := http.NewRequest("POST", s.URL+"/api/test", strings.NewReader(data))
		r.Header.Set("Content-Type", "application/json")
		res, e := s.Client().Do(r)
		if e != nil {
			t.Fatal(e)
		}
		res.Body.Close()
		if res.StatusCode != 400 {
			t.Fatal(res.StatusCode)
		}
	}
	req(t, s.Client(), s.URL, "POST", "/api/test", "", nil, 415, "Content-Type", "text/plain")
	fail := httptest.NewServer(Middleware(Handle(func(http.ResponseWriter, *http.Request) error { return errors.New("storage unavailable") })))
	defer fail.Close()
	req(t, fail.Client(), fail.URL, "GET", "/api/fail", "", nil, 503)
}
