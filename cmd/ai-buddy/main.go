package main

import (
	"context"
	"crypto/ed25519"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/limon4ik-black/AI-buddy/internal/core/transport/http/middleware"
	"github.com/limon4ik-black/AI-buddy/internal/feature/auth"
	"github.com/limon4ik-black/AI-buddy/internal/feature/auth/login"
	"github.com/limon4ik-black/AI-buddy/internal/feature/auth/logout"
	"github.com/limon4ik-black/AI-buddy/internal/feature/auth/refresh"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env, taking config from environment")
	}

	// Демо: ключи генерируются при старте, после рестарта старые токены невалидны.
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}

	issuer := auth.NewTokenIssuer(priv)
	users := auth.NewMemoryStore()
	rdb := redis.NewClient(&redis.Options{Addr: envOr("REDIS_ADDR", "localhost:6379")})
	refreshStore := auth.NewRedisRefreshStore(rdb)

	r := gin.Default()

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if clientID != "" && clientSecret != "" && redirectURL != "" {
		conf := login.NewGoogleConf(clientID, clientSecret, redirectURL)
		login.NewHandler(conf, login.NewService(conf, users, issuer, refreshStore)).Register(r)
	} else {
		log.Println("google oauth disabled: GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET/GOOGLE_REDIRECT_URL not set")
	}

	refresh.NewHandler(refresh.NewService(issuer, refreshStore)).Register(r)
	logout.NewHandler(logout.NewService(refreshStore)).Register(r)

	// Тестовая ручка входа без Google: GET /auth/test-login?user_id=42
	r.GET("/auth/test-login", func(c *gin.Context) {
		userID := int64(1)
		if rawID := c.Query("user_id"); rawID != "" {
			id, err := strconv.ParseInt(rawID, 10, 64)
			if err != nil || id <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
				return
			}
			userID = id
		}

		pair, err := issuePair(c.Request.Context(), issuer, refreshStore, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, pair)
	})

	r.GET("/test", middleware.AuthMiddleware(pub), func(c *gin.Context) {
		u, ok := middleware.FromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"user_id": u.ID,
		})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func issuePair(ctx context.Context, issuer *auth.TokenIssuer, store auth.RefreshStore, userID int64) (gin.H, error) {
	access, expiresIn, err := issuer.IssueAccess(userID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := auth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := store.Save(ctx, refreshToken, userID); err != nil {
		return nil, err
	}
	return gin.H{
		"token_type":    "Bearer",
		"algorithm":     "EdDSA",
		"access_token":  access,
		"expires_in":    expiresIn,
		"refresh_token": refreshToken,
		"user_id":       userID,
	}, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
