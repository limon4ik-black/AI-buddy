// Пакет service — правила работы с заметками.
package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

// Repo — то, что сервису нужно от хранилища. Интерфейс объявлен здесь, у потребителя.
type Repo interface {
	Create(ctx context.Context, n note.Note) (note.Note, error)
	Get(ctx context.Context, id, userID string) (note.Note, error)
	List(ctx context.Context, userID string, limit, offset int) ([]note.Note, error)
	Update(ctx context.Context, n note.Note) (note.Note, error)
	Delete(ctx context.Context, id, userID string) error
}

type Service struct {
	notes Repo
}

func New(notes Repo) *Service { return &Service{notes: notes} }
