package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

func (r *Repo) Get(ctx context.Context, id, userID string) (board.Board, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	b, err := scan(r.pool.QueryRow(ctx, `SELECT `+columns+` FROM board.boards WHERE id=$1 AND user_id=$2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return board.Board{}, board.ErrNotFound
	}
	return b, err
}
