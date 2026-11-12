package pg

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

func (r *Repo) Create(ctx context.Context, n note.Note) (note.Note, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	row := r.pool.QueryRow(ctx,
		`INSERT INTO note.notes (id, user_id, content) VALUES ($1, $2, $3) RETURNING `+columns,
		n.ID, n.UserID, n.Content)
	return scan(row)
}
