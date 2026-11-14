package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"
	_ "time/tzdata" // база часовых зон внутри бинаря: TZ=Europe/Moscow работает без tzdata в образе

	"github.com/gin-gonic/gin"
	"github.com/limon4ik-black/AI-buddy/note-service/onboarding"
	sharedconfig "github.com/limon4ik-black/AI-buddy/shared/platform/config"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "github.com/limon4ik-black/AI-buddy/note-service/docs" // регистрирует спеку из аннотаций
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/config"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/logger"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/metrics"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/middleware"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/postgres"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/security"
	boardrepo "github.com/limon4ik-black/AI-buddy/note-service/internal/features/board/repo/pg"
	boardservice "github.com/limon4ik-black/AI-buddy/note-service/internal/features/board/service"
	boardhttp "github.com/limon4ik-black/AI-buddy/note-service/internal/features/board/transport/http"
	noterepo "github.com/limon4ik-black/AI-buddy/note-service/internal/features/note/repo/pg"
	noteservice "github.com/limon4ik-black/AI-buddy/note-service/internal/features/note/service"
	notehttp "github.com/limon4ik-black/AI-buddy/note-service/internal/features/note/transport/http"
)

const defaultOrigins = "http://127.0.0.1:8080,http://localhost:8080,http://127.0.0.1:5173,http://localhost:5173"

// @title       Note Service
// @version     1.0
// @description Заметки и доски пользователя. Владелец берётся из подписанного токена и сверяется с ?user_id= из запроса.
// @BasePath    /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	ctx := context.Background()

	log, closeLog := logger.New(logger.FromEnv())
	defer closeLog()

	keys, err := security.ParseKeys("", config.Env("JWT_PUBLIC_KEY", ""))
	if err != nil {
		log.Error("jwt public key required", zap.Error(err))
		os.Exit(1)
	}

	pool, err := postgres.Open(ctx, postgres.DSN())
	if err != nil {
		log.Error("database connection failed", zap.Error(err))
		os.Exit(1)
	}
	defer pool.Close()

	timeout := config.Duration("POSTGRES_TIMEOUT", 10*time.Second)
	notes := notehttp.New(noteservice.New(noterepo.New(pool, timeout)))
	boards := boardhttp.New(boardservice.New(boardrepo.New(pool, timeout)))

	r := gin.New()
	// CORS первым: preflight уходит без Authorization, авторизация его бы отвергла.
	r.Use(middleware.Logger(log), middleware.Recover(log), middleware.CORS(strings.Split(config.Env("CORS_ORIGINS", defaultOrigins), ",")), metrics.Middleware("note-service"))
	// Без авторизации: страница должна открываться до того, как в неё ввели токен.
	// Пустой путь библиотека не отдаёт (404), поэтому «/swagger» и «/swagger/» ведём на index.html.
	swagger := ginSwagger.WrapHandler(swaggerFiles.Handler)
	r.GET("/swagger/*any", func(c *gin.Context) {
		if path := c.Param("any"); path == "" || path == "/" {
			c.Redirect(301, "/swagger/index.html")
			return
		}
		swagger(c)
	})
	api := r.Group("/api/v1", middleware.Auth(keys))
	notes.Register(api)
	boards.Register(api)
	sharedKeys, err := sharedconfig.JWT(false)
	if err != nil {
		log.Fatal("JWT configuration", zap.Error(err))
	}
	tokens, err := sharedconfig.Secrets("GATEWAY_SERVICE_TOKEN", "CHAT_SERVICE_TOKEN", "INDEXER_SERVICE_TOKEN")
	if err != nil {
		log.Fatal("service credentials", zap.Error(err))
	}
	onboardingAPI := (&onboarding.Service{DB: pool, Keys: sharedKeys, GatewayToken: tokens["GATEWAY_SERVICE_TOKEN"], ChatToken: tokens["CHAT_SERVICE_TOKEN"], IndexerToken: tokens["INDEXER_SERVICE_TOKEN"], EmployeeURL: config.Env("EMPLOYEE_URL", "http://employee:8082"), Client: &http.Client{Timeout: 25 * time.Second}}).Handler()
	r.Any("/internal/*path", gin.WrapH(onboardingAPI))
	r.GET("/api/v1/checklist", gin.WrapH(onboardingAPI))
	r.PATCH("/api/v1/checklist/tasks/:id", gin.WrapH(onboardingAPI))
	r.GET("/api/v1/information", gin.WrapH(onboardingAPI))
	r.GET("/api/v1/information/:id", gin.WrapH(onboardingAPI))

	// /metrics без авторизации и только внутри сети: его дёргает Prometheus.
	r.GET("/metrics", gin.WrapH(metrics.Handler()))

	// /healthz без авторизации: его дёргает healthcheck из compose и проверяет заодно БД.
	r.GET("/healthz", func(c *gin.Context) {
		ping, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ping); err != nil {
			response.Abort(c, err)
			return
		}
		var imported bool
		if err := pool.QueryRow(ping, "SELECT EXISTS(SELECT 1 FROM onboarding.imports WHERE name='legacy-info-v1')").Scan(&imported); err != nil || !imported {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "onboarding_not_ready"})
			return
		}
		response.JSON(c, 200, map[string]string{"status": "ok"})
	})

	port := config.Env("PORT", "8080")
	log.Info("note-service started", zap.String("address", ":"+port), zap.String("docs", "/swagger/index.html"))
	if err := r.Run(":" + port); err != nil {
		log.Error("http server failed", zap.Error(err))
		os.Exit(1)
	}
}
