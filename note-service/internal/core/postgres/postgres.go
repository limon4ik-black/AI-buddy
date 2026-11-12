// Пакет postgres — подключение к БД.
package postgres

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/config"
)

// DSN собирает строку подключения из POSTGRES_*-переменных окружения.
func DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD")),
		Host:     config.Env("POSTGRES_HOST", "127.0.0.1") + ":" + config.Env("POSTGRES_PORT", "5432"),
		Path:     "/" + os.Getenv("POSTGRES_DB"),
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

// Open создаёт пул соединений и проверяет, что БД отвечает.
func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	ping, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ping); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
