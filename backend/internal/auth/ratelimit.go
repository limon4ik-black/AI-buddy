package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
)

type attempt struct {
	Count int
	Until time.Time
}

func (a *Service) limited(fn httpapi.Endpoint) httpapi.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) error {
		a.mu.Lock()
		if a.attempts == nil {
			a.attempts = map[string]attempt{}
		}
		now := time.Now()
		for k, v := range a.attempts {
			if now.After(v.Until) {
				delete(a.attempts, k)
			}
		}
		key := strings.Split(r.RemoteAddr, ":")[0]
		v := a.attempts[key]
		if v.Until.IsZero() {
			v.Until = now.Add(time.Minute)
		}
		v.Count++
		a.attempts[key] = v
		a.mu.Unlock()
		if v.Count > 60 {
			return httpapi.Problem(429, "rate_limited", "Слишком много попыток. Подождите минуту")
		}
		return fn(w, r)
	}
}
