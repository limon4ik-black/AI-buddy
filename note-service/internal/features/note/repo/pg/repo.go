// Пакет pg — хранилище заметок в PostgreSQL.
package pg

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

// columns — один список на все выборки, чтобы Scan и RETURNING не разъезжались.
const columns = `id, user_id, content, version, created_at, updated_at`

type Repo struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

// New: timeout — предел на одну операцию к БД (POSTGRES_TIMEOUT), оборачивается в контекст.
func New(pool *pgxpool.Pool, timeout time.Duration) *Repo {
	return &Repo{pool: pool, timeout: timeout}
}

func scan(row pgx.Row) (note.Note, error) {
	var n note.Note
	err := row.Scan(&n.ID, &n.UserID, &n.Content, &n.Version, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}
