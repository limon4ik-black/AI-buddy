package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	a := &Service{}
	r := httptest.NewRequest("POST", "/auth", nil)
	r.RemoteAddr = "127.0.0.1:1"
	fn := a.limited(func(http.ResponseWriter, *http.Request) error { return nil })
	for n := 0; n < 60; n++ {
		if e := fn(httptest.NewRecorder(), r); e != nil {
			t.Fatal(e)
		}
	}
	if e := fn(httptest.NewRecorder(), r); e == nil {
		t.Fatal("Rate limit not applied")
	}
	a.attempts["old"] = attempt{1, time.Now().Add(-time.Hour)}
	_ = fn(httptest.NewRecorder(), r)
	if _, ok := a.attempts["old"]; ok {
		t.Fatal("Expired entries retained")
	}
}
