package integration

import (
	"context"
	"testing"

	chatsvc "github.com/limon4ik-black/AI-buddy/chat-service/chat"
	"github.com/limon4ik-black/AI-buddy/shared/migrations"
)

func TestNewChatMigrationPreservesLegacyTables(t *testing.T) {
	db := database(t, "chat")
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SET LOCAL search_path TO public"); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, string(migrations.Chat)); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, "INSERT INTO chats(id,user_id,title) VALUES('legacy-chat','intern-egor','Old history')"); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err = chatsvc.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	var title string
	if err = db.QueryRow(ctx, "SELECT title FROM public.chats WHERE id='legacy-chat'").Scan(&title); err != nil {
		t.Fatal(err)
	}
	if title != "Old history" {
		t.Fatal("legacy chat changed", title)
	}
	var nativeCount int
	if err = db.QueryRow(ctx, "SELECT count(*) FROM chat_v2.chats").Scan(&nativeCount); err != nil {
		t.Fatal(err)
	}
	if nativeCount != 0 {
		t.Fatal("legacy and native chat tables were mixed")
	}
}
