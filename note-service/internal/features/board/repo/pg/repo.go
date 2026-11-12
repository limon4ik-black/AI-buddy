// Пакет pg — хранилище досок в PostgreSQL.
package pg

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

// columns — один список на все выборки, чтобы Scan и RETURNING не разъезжались.
const columns = `id, user_id, title, scene, version, created_at, updated_at`

// listColumns — без сцены: список отдаёт только метаданные, сцены тяжёлые.
const listColumns = `id, user_id, title, version, created_at, updated_at`

type Repo struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

// New: timeout — предел на одну операцию к БД (POSTGRES_TIMEOUT), оборачивается в контекст.
func New(pool *pgxpool.Pool, timeout time.Duration) *Repo {
	return &Repo{pool: pool, timeout: timeout}
}

func scan(row pgx.Row) (board.Board, error) {
	var b board.Board
	err := row.Scan(&b.ID, &b.UserID, &b.Title, &b.Scene, &b.Version, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}
