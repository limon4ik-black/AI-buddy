package core_test

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/config"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/logger"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/postgres"
)

func TestConfigurationAndDatabaseConnection(t *testing.T) {
	t.Setenv("TEST_SETTING", "")
	if config.Env("TEST_SETTING", "fallback") != "fallback" {
		t.Fatal("missing fallback")
	}
	t.Setenv("TEST_SETTING", "value")
	if config.Env("TEST_SETTING", "fallback") != "value" {
		t.Fatal("missing setting")
	}
	t.Setenv("TEST_DURATION", "invalid")
	if config.Duration("TEST_DURATION", time.Second) != time.Second {
		t.Fatal("invalid duration accepted")
	}
	t.Setenv("TEST_DURATION", "250ms")
	if config.Duration("TEST_DURATION", time.Second) != 250*time.Millisecond {
		t.Fatal("duration ignored")
	}
	t.Setenv("POSTGRES_USER", "demo")
	t.Setenv("POSTGRES_PASSWORD", "special:@/?")
	t.Setenv("POSTGRES_HOST", "database")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "notes")
	u, err := url.Parse(postgres.DSN())
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	if password != "special:@/?" || u.Host != "database:5432" || u.Path != "/notes" {
		t.Fatal("incorrect DSN escaping")
	}
	pool, err := postgres.Open(context.Background(), os.Getenv("TEST_DB_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	pool.Close()
	for _, dsn := range []string{"://", "postgres://postgres@127.0.0.1:1/missing?connect_timeout=1"} {
		if p, err := postgres.Open(context.Background(), dsn); err == nil {
			p.Close()
			t.Fatal("broken connection accepted")
		}
	}
}

func TestLoggerWritesFileAndSurvivesUnavailableDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOG_DIR", dir)
	t.Setenv("LOG_LEVEL", "not-a-level")
	log, closeLog := logger.New(logger.FromEnv())
	log.Info("workspace verification")
	_ = log.Sync()
	closeLog()
	files, _ := filepath.Glob(filepath.Join(dir, "*.log"))
	if len(files) != 1 {
		t.Fatal("log file missing")
	}
	data, err := os.ReadFile(files[0])
	if err != nil || !strings.Contains(string(data), "workspace verification") {
		t.Fatal("log message missing")
	}
	blocked := filepath.Join(dir, "file")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	log, closeLog = logger.New(logger.Config{Level: "error", Dir: filepath.Join(blocked, "child")})
	log.Error("still operational")
	closeLog()
}
