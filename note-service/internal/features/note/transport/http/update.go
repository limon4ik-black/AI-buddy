package http

import (
	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

// updateRequest — тело PATCH /api/v1/notes/{id}.
type updateRequest struct {
	Content string `json:"content"`
	Version int    `json:"version"`
}

// @Summary      Изменить текст заметки
// @Description  Оптимистичная блокировка: version должен совпадать с текущим в базе.
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        id      path  string        true "id заметки"
// @Param        user_id query string        false "необязательно; если указан, должен совпадать с sub токена"
// @Param        body    body  updateRequest true "новый текст и ожидаемая версия"
// @Success      200 {object} note.Note
// @Failure      400 {object} response.ErrorBody "некорректный запрос или нет version"
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Failure      404 {object} response.ErrorBody "заметки нет или она чужая"
// @Failure      409 {object} response.ErrorBody "версия устарела, в тексте актуальная"
// @Security     BearerAuth
// @Router       /notes/{id} [patch]
func (t *Transport) update(c *gin.Context) {
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
	var in updateRequest
	if err := response.Decode(c, &in); err != nil {
		abort(c, err)
		return
	}
	if in.Version < 1 {
		abort(c, response.Invalid("Нужна версия заметки — version"))
		return
	}
	v, err := t.notes.Update(c.Request.Context(), a, note.Note{ID: id, Content: in.Content, Version: in.Version})
	if err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 200, v)
}
