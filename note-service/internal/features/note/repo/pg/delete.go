package pg

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

// Delete убирает заметку владельца; чужая и несуществующая неотличимы.
func (r *Repo) Delete(ctx context.Context, id, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	tag, err := r.pool.Exec(ctx, `DELETE FROM note.notes WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return note.ErrNotFound
	}
	return nil
}
