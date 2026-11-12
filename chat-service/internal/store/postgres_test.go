package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
)

func TestPostgres(t *testing.T) {
	dsn := os.Getenv("CHAT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set CHAT_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := "chat_test_" + uuid.New().String()[:8]
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := &Postgres{Pool: pool}
	for i := 0; i < 2; i++ {
		if err = repo.Migrate(ctx); err != nil {
			t.Fatal(err)
		}
	}
	u, other := uuid.New(), uuid.New()
	one, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	two, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	h, err := repo.History(ctx, u, one.ID)
	if err != nil || h == nil || len(h) != 0 {
		t.Fatal(h, err)
	}
	first, err := repo.Append(ctx, u, one.ID, "user", "Привет")
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.Append(ctx, u, one.ID, "assistant", "Ответ")
	if err != nil {
		t.Fatal(err)
	}
	h, err = repo.History(ctx, u, one.ID)
	if err != nil || len(h) != 2 || h[0].ID != first.ID || h[1].ID != second.ID {
		t.Fatal(h, err)
	}
	list, err := repo.List(ctx, u)
	if err != nil || len(list) != 2 || list[0].ID != one.ID || list[1].ID != two.ID {
		t.Fatal(list, err)
	}
	list, err = repo.List(ctx, other)
	if err != nil || list == nil || len(list) != 0 {
		t.Fatal(list, err)
	}
	if _, err = repo.History(ctx, other, one.ID); !errors.Is(err, chat.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err = repo.Append(ctx, other, one.ID, "user", "blocked"); !errors.Is(err, chat.ErrNotFound) {
		t.Fatal(err)
	}
	if err = repo.Delete(ctx, other, one.ID); !errors.Is(err, chat.ErrNotFound) {
		t.Fatal(err)
	}
	if err = repo.Delete(ctx, u, one.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM messages WHERE chat_id=$1", one.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal(count, err)
	}
	if _, err = repo.History(ctx, u, one.ID); !errors.Is(err, chat.ErrNotFound) {
		t.Fatal(err)
	}
	if err = repo.Delete(ctx, u, one.ID); !errors.Is(err, chat.ErrNotFound) {
		t.Fatal(err)
	}
	pool.Close()
	if _, err = repo.List(ctx, u); err == nil {
		t.Fatal("expected unavailable database error")
	}
}
