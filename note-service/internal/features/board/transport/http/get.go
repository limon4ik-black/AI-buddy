package http

import (
	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
)

// @Summary      Получить доску
// @Description  Отдаёт доску целиком, вместе со сценой. Владелец берётся из токена.
// @Tags         boards
// @Produce      json
// @Param        id      path  string true "id доски"
// @Param        user_id query string false "необязательно; если указан, должен совпадать с sub токена"
// @Success      200 {object} board.Board
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Failure      404 {object} response.ErrorBody "доски нет или она чужая"
// @Security     BearerAuth
// @Router       /boards/{id} [get]
func (t *Transport) get(c *gin.Context) {
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
	b, err := t.boards.Get(c.Request.Context(), a, id)
	if err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 200, b)
}
