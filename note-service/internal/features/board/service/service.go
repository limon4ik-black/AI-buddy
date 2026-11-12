// Пакет service — правила работы с досками.
package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

// Repo — то, что сервису нужно от хранилища. Интерфейс объявлен здесь, у потребителя.
type Repo interface {
	Create(ctx context.Context, b board.Board) (board.Board, error)
	Get(ctx context.Context, id, userID string) (board.Board, error)
	List(ctx context.Context, userID string, limit, offset int) ([]board.Board, error)
	Update(ctx context.Context, b board.Board) (board.Board, error)
	Delete(ctx context.Context, id, userID string) error
}

type Service struct {
	boards Repo
}

func New(boards Repo) *Service { return &Service{boards: boards} }
