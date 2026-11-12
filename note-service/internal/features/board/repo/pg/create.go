package pg

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

func (r *Repo) Create(ctx context.Context, b board.Board) (board.Board, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	row := r.pool.QueryRow(ctx,
		`INSERT INTO board.boards (id, user_id, title, scene) VALUES ($1, $2, $3, $4) RETURNING `+columns,
		b.ID, b.UserID, b.Title, b.Scene)
	return scan(row)
}
