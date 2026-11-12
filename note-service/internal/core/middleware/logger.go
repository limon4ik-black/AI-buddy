// Пакет middleware — общие gin-мидлвари сервиса.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger пишет строку на каждый запрос: 5xx и накопленные ошибки — уровнем Error.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path, query := c.Request.URL.Path, c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.Int("size", c.Writer.Size()),
			zap.String("ip", c.ClientIP()),
		}
		switch {
		case len(c.Errors) > 0:
			log.Error("request failed", append(fields, zap.String("error", c.Errors.String()))...)
		case status >= 500:
			log.Error("request", fields...)
		default:
			log.Info("request", fields...)
		}
	}
}
