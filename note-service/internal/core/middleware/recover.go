package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
)

// Recover превращает панику в 500 и оставляет в логе стек.
func Recover(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Stack("stack"),
				)
				response.Abort(c, response.Problem(500, "internal", "Внутренняя ошибка сервиса"))
			}
		}()
		c.Next()
	}
}
