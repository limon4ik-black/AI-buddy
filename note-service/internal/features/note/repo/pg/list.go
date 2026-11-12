package pg

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

// List отдаёт заметки владельца, свежие сверху.
func (r *Repo) List(ctx context.Context, userID string, limit, offset int) ([]note.Note, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	rows, err := r.pool.Query(ctx,
		`SELECT `+columns+` FROM note.notes WHERE user_id=$1 ORDER BY updated_at DESC, id DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]note.Note, 0, limit) // не nil: пустой список уезжает как [], а не null
	for rows.Next() {
		n, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, rows.Err()
}
