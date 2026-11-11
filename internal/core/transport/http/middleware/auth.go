package middleware

import (
	"crypto/ed25519"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware достает JWT из заголовка Authorization (Bearer <токен>),
// проверяет подпись публичным ключом и кладет пользователя в контекст запроса.
//
// Поддерживается только EdDSA (Ed25519) - тот же алгоритм, которым токены выдаются.
// Проверка асимметричная: мидлваре нужен только ПУБЛИЧНЫЙ ключ, подделать токен
// им нельзя (для этого нужен приватный ключ, который есть лишь у того, кто выдает токены).
func AuthMiddleware(pub ed25519.PublicKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		parsed, err := jwt.NewParser(jwt.WithValidMethods([]string{"EdDSA"})).
			Parse(token, func(*jwt.Token) (any, error) { return pub, nil })
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		id, err := userIDFromClaims(parsed.Claims.(jwt.MapClaims))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Кладем через stdlib-контекст: хендлеры не зависят от gin.
		c.Request = c.Request.WithContext(NewContext(c.Request.Context(), User{ID: id}))
		c.Next()
	}
}

// userIDFromClaims вытаскивает claim user_id. JSON-число приходит как float64,
// но допускаем и строку на случай, если токен выдали строкой.
func userIDFromClaims(claims jwt.MapClaims) (int64, error) {
	switch id := claims["user_id"].(type) {
	case float64:
		return int64(id), nil
	case string:
		return strconv.ParseInt(id, 10, 64)
	default:
		return 0, errors.New("missing user_id claim")
	}
}

// bearerToken разбирает заголовок формата "Bearer <токен>".
func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}
