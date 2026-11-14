package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/limon4ik-black/AI-buddy/shared/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/shared/platform/security"
)

type contextKey string

const userKey contextKey = "user"

const requestKey contextKey = "requestId"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if len(id) != 32 {
			id = idgen.New()
		}
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		// Дальше идёт копия запроса, но шаблон маршрута переносим обратно:
		// ServeMux проставляет r.Pattern уже копии, а мидлварь метрик читает
		// его снаружи — без этого метка пути была бы пустой («unmatched»).
		inner := r.WithContext(context.WithValue(r.Context(), requestKey, id))
		next.ServeHTTP(w, inner)
		r.Pattern = inner.Pattern
	})
}

func ServiceToken(secret string, fn Endpoint) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) error {
		if !security.Equal(secret, r.Header.Get("X-Service-Token")) {
			return Problem(401, "unauthorized_service", "Нужны сервисные полномочия")
		}
		return fn(w, r)
	}
}

func Authenticated(keys *security.Keys, fn Endpoint) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) error {
		claims, err := keys.Verify(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if err != nil {
			return Problem(401, "unauthorized", "Войдите в систему")
		}
		return fn(w, r.WithContext(context.WithValue(r.Context(), userKey, claims)))
	}
}

func User(r *http.Request) security.Claims { return r.Context().Value(userKey).(security.Claims) }

// RequestID returns the correlation ID installed by Middleware.
func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestKey).(string)
	return value
}
