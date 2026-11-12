package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

// Update меняет название и сцену целиком, если версия в базе совпадает с ожидаемой.
func (r *Repo) Update(ctx context.Context, b board.Board) (board.Board, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	row := r.pool.QueryRow(ctx,
		`UPDATE board.boards SET title=$1, scene=$2, version=version+1, updated_at=now()
		 WHERE id=$3 AND user_id=$4 AND version=$5 RETURNING `+columns,
		b.Title, b.Scene, b.ID, b.UserID, b.Version)
	updated, err := scan(row)
	if err == nil {
		return updated, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return board.Board{}, err
	}
	return board.Board{}, r.reason(ctx, b.ID, b.UserID)
}

// reason отвечает, почему UPDATE не задел строк: доски нет или версия устарела.
func (r *Repo) reason(ctx context.Context, id, userID string) error {
	var current int
	err := r.pool.QueryRow(ctx, `SELECT version FROM board.boards WHERE id=$1 AND user_id=$2`, id, userID).Scan(&current)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return board.ErrNotFound
	case err != nil:
		return err
	}
	return &board.ConflictError{Version: current}
}
