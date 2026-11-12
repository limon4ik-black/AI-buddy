// Пакет http — транспорт сервиса-блокнота: ручки на gin.
package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

// Service — то, что транспорту нужно от бизнес-логики. Интерфейс объявлен здесь,
// у потребителя: пакет service о нём не знает.
type Service interface {
	Create(ctx context.Context, a note.Access, content string) (note.Note, error)
	Get(ctx context.Context, a note.Access, id string) (note.Note, error)
	List(ctx context.Context, a note.Access, limit, offset int) ([]note.Note, error)
	Update(ctx context.Context, a note.Access, n note.Note) (note.Note, error)
	Delete(ctx context.Context, a note.Access, id string) error
}

type Transport struct {
	notes Service
}

func New(notes Service) *Transport { return &Transport{notes: notes} }

// Register вешает ручки сервиса на группу /api/v1.
func (t *Transport) Register(api *gin.RouterGroup) {
	api.POST("/notes", t.create)
	api.GET("/notes", t.list)
	api.GET("/notes/:id", t.get)
	api.PATCH("/notes/:id", t.update)
	api.DELETE("/notes/:id", t.delete)
}
