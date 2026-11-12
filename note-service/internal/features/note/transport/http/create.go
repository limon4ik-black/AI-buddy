package http

import (
	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
)

// createRequest — тело POST /api/v1/notes.
type createRequest struct {
	Content string `json:"content"`
}

// @Summary      Создать заметку
// @Description  Владелец берётся из токена, ?user_id= должен совпасть с ним.
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        user_id query string        false "необязательно; если указан, должен совпадать с sub токена"
// @Param        body    body  createRequest true "текст заметки"
// @Success      201 {object} note.Note
// @Failure      400 {object} response.ErrorBody "некорректный запрос"
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Security     BearerAuth
// @Router       /notes [post]
func (t *Transport) create(c *gin.Context) {
	a, err := owner(c)
	if err != nil {
		abort(c, err)
		return
	}
	var in createRequest
	if err := response.Decode(c, &in); err != nil {
		abort(c, err)
		return
	}
	v, err := t.notes.Create(c.Request.Context(), a, in.Content)
	if err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 201, v)
}
