package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/security"
)

const userKey = "userID"

// Auth проверяет подписанный токен и кладёт владельца в контекст.
// Владелец берётся только из подписанного JWT.
func Auth(keys *security.Keys) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := keys.Verify(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if err != nil {
			response.Abort(c, response.Problem(401, "unauthorized", "Войдите в систему"))
			return
		}
		c.Set(userKey, claims.Subject)
		c.Next()
	}
}

// UserID — владелец запроса, которого поставила Auth.
func UserID(c *gin.Context) string { return c.GetString(userKey) }
