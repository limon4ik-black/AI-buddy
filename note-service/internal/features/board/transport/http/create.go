package http

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

// createRequest — тело POST /api/v1/boards.
type createRequest struct {
	Title string          `json:"title"`
	Scene json.RawMessage `json:"scene" swaggertype:"object"`
}

// @Summary      Создать доску
// @Description  Владелец берётся из токена, ?user_id= должен совпасть с ним. Сцена передаётся целиком.
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        user_id query string        false "необязательно; если указан, должен совпадать с sub токена"
// @Param        body    body  createRequest true "название и сцена Excalidraw"
// @Success      201 {object} board.Board
// @Failure      400 {object} response.ErrorBody "некорректный запрос или сцена"
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Security     BearerAuth
// @Router       /boards [post]
func (t *Transport) create(c *gin.Context) {
	a, err := owner(c)
	if err != nil {
		abort(c, err)
		return
	}
	var in createRequest
	// Лимит выше, чем у заметок: сцена холста в разы тяжелее текста.
	if err := response.DecodeLimit(c, &in, board.MaxScene+maxOverhead); err != nil {
		abort(c, err)
		return
	}
	v, err := t.boards.Create(c.Request.Context(), a, in.Title, in.Scene)
	if err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 201, v)
}
