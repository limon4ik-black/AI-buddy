package chat_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/testutil"
)

func TestSendHistoryAndFailure(t *testing.T) {
	ctx := context.Background()
	repo := testutil.New()
	u := uuid.New()
	c, _ := repo.Create(ctx, u)
	for i := 0; i < 22; i++ {
		_, _ = repo.Append(ctx, u, c.ID, "user", fmt.Sprint(i))
	}
	svc := &chat.Service{Repository: repo, AI: testutil.AIFunc(func(_ context.Context, message string, h []chat.Turn) (string, error) {
		if message != "current" || len(h) != 20 || h[0].Content != "2" || h[19].Content != "21" {
			t.Fatalf("wrong AI context: %q %+v", message, h)
		}
		return "answer", nil
	})}
	answer, err := svc.Send(ctx, u, c.ID, " current ")
	if err != nil || answer != "answer" {
		t.Fatal(answer, err)
	}
	history, _ := repo.History(ctx, u, c.ID)
	if len(history) != 24 || history[23].Role != "assistant" {
		t.Fatal(history)
	}
	svc.AI = testutil.AIFunc(func(context.Context, string, []chat.Turn) (string, error) { return "", errors.New("secret") })
	_, err = svc.Send(ctx, u, c.ID, "keep me")
	if !errors.Is(err, chat.ErrAI) {
		t.Fatal(err)
	}
	history, _ = repo.History(ctx, u, c.ID)
	if len(history) != 25 || history[24].Content != "keep me" {
		t.Fatal(history)
	}
}
func TestSendValidationAndOwnership(t *testing.T) {
	ctx := context.Background()
	repo := testutil.New()
	u := uuid.New()
	c, _ := repo.Create(ctx, u)
	svc := &chat.Service{Repository: repo, AI: testutil.AIFunc(func(context.Context, string, []chat.Turn) (string, error) {
		t.Fatal("AI should not be called")
		return "", nil
	})}
	for _, s := range []string{"", " \n\t", strings.Repeat("я", 8001)} {
		if _, err := svc.Send(ctx, u, c.ID, s); !errors.Is(err, chat.ErrMessage) {
			t.Fatal(err)
		}
	}
	for _, id := range []uuid.UUID{c.ID, uuid.New()} {
		if _, err := svc.Send(ctx, uuid.New(), id, "hello"); !errors.Is(err, chat.ErrNotFound) {
			t.Fatal(err)
		}
	}
	h, _ := repo.History(ctx, u, c.ID)
	if len(h) != 0 {
		t.Fatal(h)
	}
}
