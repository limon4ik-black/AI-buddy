package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// testRouter собирает роутер как в бою: сначала владелец в контексте (это делает
// Auth), потом лимитер. Имя владельца берём из заголовка, чтобы в одном роутере
// проверить разных клиентов.
func testRouter(rps float64, burst int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if id := c.GetHeader("X-Test-User"); id != "" {
			c.Set(userKey, id)
		}
		c.Next()
	}, RateLimit(rps, burst))
	r.GET("/api/v1/notes", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// hit делает один запрос от указанного клиента (или с указанного адреса).
func hit(r *gin.Engine, user, ip string) (int, http.Header) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notes", nil)
	req.RemoteAddr = ip + ":1234"
	if user != "" {
		req.Header.Set("X-Test-User", user)
	}
	r.ServeHTTP(w, req)
	return w.Code, w.Header()
}

// Пачка проходит целиком, следующий запрос получает 429 с Retry-After.
// Пополнение тут почти нулевое, чтобы отказ был именно от исчерпания пачки.
func TestRateLimitStopsAfterBurst(t *testing.T) {
	r := testRouter(0.0001, 3)
	for i := 1; i <= 3; i++ {
		if code, _ := hit(r, "anton", "10.0.0.1"); code != http.StatusOK {
			t.Fatalf("запрос %d из пачки получил %d, ждали 200", i, code)
		}
	}
	code, h := hit(r, "anton", "10.0.0.1")
	if code != http.StatusTooManyRequests {
		t.Fatalf("запрос сверх пачки получил %d, ждали 429", code)
	}
	if h.Get("Retry-After") == "" {
		t.Fatal("в отказе нет заголовка Retry-After")
	}
}

// Квота у каждого своя: исчерпанная у одного не мешает другому.
func TestRateLimitIsPerUser(t *testing.T) {
	r := testRouter(0.0001, 1)
	if code, _ := hit(r, "anton", "10.0.0.1"); code != http.StatusOK {
		t.Fatalf("первый запрос anton получил %d", code)
	}
	if code, _ := hit(r, "anton", "10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("второй запрос anton получил %d, ждали 429", code)
	}
	if code, _ := hit(r, "ilvina", "10.0.0.1"); code != http.StatusOK {
		t.Fatalf("запрос другого пользователя получил %d, ждали 200 — квота должна быть персональной", code)
	}
}

// Без токена (владельца в контексте нет) ключом становится адрес клиента.
func TestRateLimitFallsBackToIP(t *testing.T) {
	r := testRouter(0.0001, 1)
	if code, _ := hit(r, "", "10.0.0.1"); code != http.StatusOK {
		t.Fatalf("первый запрос с адреса получил %d", code)
	}
	if code, _ := hit(r, "", "10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("второй запрос с того же адреса получил %d, ждали 429", code)
	}
	if code, _ := hit(r, "", "10.0.0.2"); code != http.StatusOK {
		t.Fatalf("запрос с другого адреса получил %d, ждали 200", code)
	}
}

// После паузы квота восстанавливается: 100 в секунду — за 50 мс вернётся пять.
func TestRateLimitRefillsOverTime(t *testing.T) {
	r := testRouter(100, 1)
	if code, _ := hit(r, "anton", "10.0.0.1"); code != http.StatusOK {
		t.Fatalf("первый запрос получил %d", code)
	}
	if code, _ := hit(r, "anton", "10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("второй запрос получил %d, ждали 429", code)
	}
	time.Sleep(50 * time.Millisecond)
	if code, _ := hit(r, "anton", "10.0.0.1"); code != http.StatusOK {
		t.Fatalf("после паузы запрос получил %d, ждали 200 — квота должна восстанавливаться", code)
	}
}

// Простаивающий клиент выбрасывается из таблицы: иначе она растёт бесконечно.
// Проверяем по поведению — после вытеснения у клиента снова полная пачка, хотя
// пополнение почти нулевое и само по себе токен вернуть не могло.
func TestRateLimitEvictsIdleClients(t *testing.T) {
	idleTTL, sweepEvery = time.Millisecond, time.Millisecond
	defer func() { idleTTL, sweepEvery = 10*time.Minute, time.Minute }()

	r := testRouter(0.0001, 1)
	if code, _ := hit(r, "anton", "10.0.0.1"); code != http.StatusOK {
		t.Fatalf("первый запрос получил %d", code)
	}
	if code, _ := hit(r, "anton", "10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("второй запрос получил %d, ждали 429", code)
	}
	time.Sleep(5 * time.Millisecond)
	code, _ := hit(r, "anton", "10.0.0.1")
	if code != http.StatusOK {
		t.Fatalf("после простоя запрос получил %d, ждали 200 — запись не вытеснилась и таблица будет расти", code)
	}
}

// Пределы читаются из окружения, а мусор в переменной не ломает сервис.
func TestRateLimitFromEnv(t *testing.T) {
	env := func(values map[string]string) func(string, string) string {
		return func(key, def string) string {
			if v, ok := values[key]; ok {
				return v
			}
			return def
		}
	}
	if m := RateLimitFromEnv(env(map[string]string{"RATE_LIMIT_RPS": "1", "RATE_LIMIT_BURST": "1"})); m == nil {
		t.Fatal("мидлварь не собралась из переменных окружения")
	}
	if m := RateLimitFromEnv(env(map[string]string{"RATE_LIMIT_RPS": "ерунда"})); m == nil {
		t.Fatal("мусор в переменной сломал сборку мидлвари")
	}
}
