// Пакет http — транспорт сервиса досок: ручки на gin.
package http

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

// Service — то, что транспорту нужно от бизнес-логики. Интерфейс объявлен здесь,
// у потребителя: пакет service о нём не знает.
type Service interface {
	Create(ctx context.Context, a board.Access, title string, scene json.RawMessage) (board.Board, error)
	Get(ctx context.Context, a board.Access, id string) (board.Board, error)
	List(ctx context.Context, a board.Access, limit, offset int) ([]board.Board, error)
	Update(ctx context.Context, a board.Access, b board.Board) (board.Board, error)
	Delete(ctx context.Context, a board.Access, id string) error
}

type Transport struct {
	boards Service
}

func New(boards Service) *Transport { return &Transport{boards: boards} }

// Register вешает ручки сервиса на группу /api/v1.
func (t *Transport) Register(api *gin.RouterGroup) {
	api.POST("/boards", t.create)
	api.GET("/boards", t.list)
	api.GET("/boards/:id", t.get)
	api.PATCH("/boards/:id", t.update)
	api.DELETE("/boards/:id", t.delete)
}
