package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

// Update меняет текст, если версия в базе совпадает с ожидаемой.
func (r *Repo) Update(ctx context.Context, n note.Note) (note.Note, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	row := r.pool.QueryRow(ctx,
		`UPDATE note.notes SET content=$1, version=version+1, updated_at=now()
		 WHERE id=$2 AND user_id=$3 AND version=$4 RETURNING `+columns,
		n.Content, n.ID, n.UserID, n.Version)
	updated, err := scan(row)
	if err == nil {
		return updated, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return note.Note{}, err
	}
	return note.Note{}, r.reason(ctx, n.ID, n.UserID)
}

// reason отвечает, почему UPDATE не задел строк: заметки нет или версия устарела.
func (r *Repo) reason(ctx context.Context, id, userID string) error {
	var current int
	err := r.pool.QueryRow(ctx, `SELECT version FROM note.notes WHERE id=$1 AND user_id=$2`, id, userID).Scan(&current)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return note.ErrNotFound
	case err != nil:
		return err
	}
	return &note.ConflictError{Version: current}
}
