package http

import (
	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
)

// @Summary      Получить заметку
// @Description  Владелец берётся из токена, ?user_id= должен совпасть с ним.
// @Tags         notes
// @Produce      json
// @Param        id      path  string true "id заметки"
// @Param        user_id query string false "необязательно; если указан, должен совпадать с sub токена"
// @Success      200 {object} note.Note
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Failure      404 {object} response.ErrorBody "заметки нет или она чужая"
// @Security     BearerAuth
// @Router       /notes/{id} [get]
func (t *Transport) get(c *gin.Context) {
	a, err := owner(c)
	if err != nil {
		abort(c, err)
		return
	}
	id, err := noteID(c)
	if err != nil {
		abort(c, err)
		return
	}
	n, err := t.notes.Get(c.Request.Context(), a, id)
	if err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 200, n)
}
