// Пакет metrics — gin-адаптер к общей реализации метрик (shared/platform/metrics):
// считалки одни на всю репу, здесь только способ их вызова из gin.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	sharedmetrics "github.com/limon4ik-black/AI-buddy/shared/platform/metrics"
)

// Handler отдаёт метрики в формате Prometheus.
func Handler() http.Handler {
	return sharedmetrics.Handler()
}

// Middleware считает запросы. В label path идёт шаблон маршрута (gin FullPath),
// а не сырой URL: иначе каждый id заводил бы новую серию метрик и Prometheus бы задохнулся.
func Middleware(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Свои скрейпы и опросы здоровья не считаем: иначе в графиках
		// всегда висит постоянная линия от них.
		if path := c.FullPath(); path == "/metrics" || path == "/healthz" {
			c.Next()
			return
		}
		done := sharedmetrics.InFlight(service)
		start := time.Now()

		c.Next()

		done()
		sharedmetrics.Observe(service, c.Request.Method, c.FullPath(), strconv.Itoa(c.Writer.Status()), time.Since(start).Seconds())
	}
}
