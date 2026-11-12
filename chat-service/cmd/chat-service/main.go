package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/ai"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/store"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/transport"
)

func main() {
	if err := run(); err != nil {
		slog.Error("chat-service stopped", "reason", err.Error())
		os.Exit(1)
	}
}
func duration(name, fallback string) (time.Duration, error) {
	v := os.Getenv(name)
	if v == "" {
		v = fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return 0, errors.New("invalid " + name)
	}
	return d, nil
}
func run() error {
	databaseURL, aiURL, token := os.Getenv("DATABASE_URL"), os.Getenv("AI_SERVICE_URL"), os.Getenv("FRIEND_TOKEN")
	if databaseURL == "" || aiURL == "" || strings.TrimSpace(token) == "" {
		return errors.New("DATABASE_URL, AI_SERVICE_URL and FRIEND_TOKEN are required")
	}
	u, err := url.Parse(aiURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("invalid AI_SERVICE_URL")
	}
	aiTimeout, err := duration("AI_HTTP_TIMEOUT", "60s")
	if err != nil {
		return err
	}
	timeout, err := duration("REQUEST_TIMEOUT", "75s")
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, 15*time.Second)
	pool, err := pgxpool.New(startup, databaseURL)
	if err != nil {
		cancel()
		return errors.New("invalid database configuration")
	}
	defer pool.Close()
	repository := &store.Postgres{Pool: pool}
	err = pool.Ping(startup)
	if err == nil {
		err = repository.Migrate(startup)
	}
	cancel()
	if err != nil {
		return errors.New("database initialization failed")
	}
	client := &http.Client{Timeout: aiTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	handler := &transport.Handler{Service: &chat.Service{Repository: repository, AI: &ai.Client{URL: aiURL, Token: token, HTTP: client}}, Timeout: timeout}
	if origins := os.Getenv("WS_ORIGINS"); origins != "" {
		for _, origin := range strings.Split(origins, ",") {
			handler.Origins = append(handler.Origins, strings.TrimSpace(origin))
		}
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	server := &http.Server{Addr: addr, Handler: handler.Routes(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, BaseContext: func(net.Listener) context.Context { return ctx }}
	result := make(chan error, 1)
	go func() { slog.Info("chat-service listening", "addr", addr); result <- server.ListenAndServe() }()
	select {
	case err = <-result:
		if !errors.Is(err, http.ErrServerClosed) {
			return errors.New("HTTP server failed")
		}
		return nil
	case <-ctx.Done():
		shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		if err = server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return errors.New("HTTP shutdown timed out")
		}
		drained := make(chan struct{})
		go func() { handler.Wait(); close(drained) }()
		select {
		case <-drained:
		case <-shutdown.Done():
			return errors.New("WebSocket shutdown timed out")
		}
		return nil
	}
}
