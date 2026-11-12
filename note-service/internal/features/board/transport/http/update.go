package http

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

// updateRequest — тело PATCH /api/v1/boards/{id}.
type updateRequest struct {
	Title   string          `json:"title"`
	Scene   json.RawMessage `json:"scene" swaggertype:"object"`
	Version int             `json:"version"`
}

// @Summary      Изменить доску
// @Description  Заменяет название и сцену целиком. Оптимистичная блокировка: version должен совпадать с текущим в базе.
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        id      path  string        true "id доски"
// @Param        user_id query string        false "необязательно; если указан, должен совпадать с sub токена"
// @Param        body    body  updateRequest true "новое название, сцена и ожидаемая версия"
// @Success      200 {object} board.Board
// @Failure      400 {object} response.ErrorBody "некорректный запрос, сцена или нет version"
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Failure      404 {object} response.ErrorBody "доски нет или она чужая"
// @Failure      409 {object} response.ErrorBody "версия устарела, в тексте актуальная"
// @Security     BearerAuth
// @Router       /boards/{id} [patch]
func (t *Transport) update(c *gin.Context) {
	a, err := owner(c)
	if err != nil {
		abort(c, err)
		return
	}
	id, err := boardID(c)
	if err != nil {
		abort(c, err)
		return
	}
	var in updateRequest
	if err := response.DecodeLimit(c, &in, board.MaxScene+maxOverhead); err != nil {
		abort(c, err)
		return
	}
	if in.Version < 1 {
		abort(c, response.Invalid("Нужна версия доски — version"))
		return
	}
	v, err := t.boards.Update(c.Request.Context(), a, board.Board{ID: id, Title: in.Title, Scene: in.Scene, Version: in.Version})
	if err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 200, v)
}
