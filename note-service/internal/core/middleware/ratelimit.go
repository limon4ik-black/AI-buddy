package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
)

// Пределы по умолчанию: 20 запросов в секунду, пачкой до 40. Интерфейсу хватает
// с запасом, а цикл в коде или чужой скрипт об это спотыкается.
const (
	defaultRPS   = 20
	defaultBurst = 40
)

// Настройки вытеснения. Держать лимитер клиента, который больше не приходит,
// незачем: без вытеснения таблица росла бы вместе с числом уникальных клиентов.
// Переменные, а не константы, чтобы тест мог ужать их до миллисекунд.
var (
	idleTTL    = 10 * time.Minute
	sweepEvery = time.Minute
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimit пускает не больше rps запросов в секунду, пачкой до burst.
// Ключ — владелец из токена (его ставит Auth), а если токена нет — адрес клиента:
// за прокси у всех клиентов один IP, и считать их одной квотой было бы неверно.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 {
		rps = defaultRPS
	}
	if burst <= 0 {
		burst = defaultBurst
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*limiterEntry)
		sweptAt = time.Now()
	)
	return func(c *gin.Context) {
		key := clientKey(c)
		now := time.Now()

		mu.Lock()
		// Чистим не на каждом запросе: раз в sweepEvery обходим таблицу и
		// выбрасываем тех, кого не видели idleTTL.
		if now.Sub(sweptAt) >= sweepEvery {
			for k, e := range clients {
				if now.Sub(e.lastSeen) >= idleTTL {
					delete(clients, k)
				}
			}
			sweptAt = now
		}
		e, ok := clients[key]
		if !ok {
			e = &limiterEntry{limiter: rate.NewLimiter(rate.Limit(rps), burst)}
			clients[key] = e
		}
		e.lastSeen = now
		allowed := e.limiter.Allow()
		mu.Unlock()

		if !allowed {
			// Retry-After в секундах: клиенту видно, через сколько возвращаться.
			c.Header("Retry-After", "1")
			response.Abort(c, response.Problem(http.StatusTooManyRequests, "rate_limited", "Слишком много запросов — сбавьте темп"))
			return
		}
		c.Next()
	}
}

// RateLimitFromEnv берёт пределы из окружения: RATE_LIMIT_RPS и RATE_LIMIT_BURST.
// get — это config.Env; значения по умолчанию подставляются, если переменной нет.
func RateLimitFromEnv(get func(string, string) string) gin.HandlerFunc {
	rps, err := strconv.ParseFloat(get("RATE_LIMIT_RPS", ""), 64)
	if err != nil {
		rps = defaultRPS
	}
	burst, err := strconv.Atoi(get("RATE_LIMIT_BURST", ""))
	if err != nil {
		burst = defaultBurst
	}
	return RateLimit(rps, burst)
}

// clientKey — по кому считать квоту.
func clientKey(c *gin.Context) string {
	if id := UserID(c); id != "" {
		return "user:" + id
	}
	return "ip:" + c.ClientIP()
}
