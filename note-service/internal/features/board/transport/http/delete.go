package http

import (
	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
)

// delete не сверяет version: удаляем то, что владелец решил удалить,
// повторный запрос просто получит 404.
// @Summary      Удалить доску
// @Description  Без проверки version: повторный запрос вернёт 404.
// @Tags         boards
// @Param        id      path  string true "id доски"
// @Param        user_id query string false "необязательно; если указан, должен совпадать с sub токена"
// @Success      204 "Удалено"
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Failure      404 {object} response.ErrorBody "доски нет, она чужая или уже удалена"
// @Security     BearerAuth
// @Router       /boards/{id} [delete]
func (t *Transport) delete(c *gin.Context) {
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
	if err := t.boards.Delete(c.Request.Context(), a, id); err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 204, nil)
}
