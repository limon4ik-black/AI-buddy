package pg

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

// List отдаёт доски владельца, свежие сверху. Сцена не выбирается: списку нужны
// только метаданные, иначе открытие вкладки выкачивало бы все холсты пользователя.
func (r *Repo) List(ctx context.Context, userID string, limit, offset int) ([]board.Board, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	rows, err := r.pool.Query(ctx,
		`SELECT `+listColumns+` FROM board.boards WHERE user_id=$1 ORDER BY updated_at DESC, id DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]board.Board, 0, limit) // не nil: пустой список уезжает как [], а не null
	for rows.Next() {
		var b board.Board
		if err := rows.Scan(&b.ID, &b.UserID, &b.Title, &b.Version, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	return items, rows.Err()
}
