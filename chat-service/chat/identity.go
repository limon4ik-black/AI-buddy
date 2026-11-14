package chat

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/store"
)

// owner preserves native UUID owners and maps signed string subjects consistently.
func owner(subject string) string {
	if id, err := uuid.Parse(subject); err == nil && len(subject) == 36 {
		return id.String()
	}
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("ai-buddy:user:"+subject)).String()
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return (&store.Postgres{Pool: pool}).Migrate(ctx)
}
