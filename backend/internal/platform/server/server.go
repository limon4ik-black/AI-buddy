package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
)

// Factory initializes one application and returns its cleanup function.
type Factory func(context.Context) (http.Handler, func(), error)

// Run owns the HTTP process lifecycle. Application wiring stays with each service.
func Run(name string, build Factory) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	handler, cleanup, err := build(ctx)
	if err != nil {
		slog.Error("application initialization failed", "service", name, "error", err)
		os.Exit(1)
	}
	defer cleanup()
	srv := &http.Server{Addr: config.Env("LISTEN_ADDR", ":8080"), Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 45 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	slog.Info("service started", "service", name, "address", srv.Addr)
	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server failed", "service", name, "error", err)
		os.Exit(1)
	}
}
