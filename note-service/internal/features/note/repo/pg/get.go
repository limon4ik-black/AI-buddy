package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

func (r *Repo) Get(ctx context.Context, id, userID string) (note.Note, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	n, err := scan(r.pool.QueryRow(ctx, `SELECT `+columns+` FROM note.notes WHERE id=$1 AND user_id=$2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return note.Note{}, note.ErrNotFound
	}
	return n, err
}
