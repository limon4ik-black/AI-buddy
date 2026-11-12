package http

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// listQuery — параметры GET /api/v1/notes.
type listQuery struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// listResponse — тело ответа GET /api/v1/notes.
type listResponse struct {
	Items  []note.Note `json:"items"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// @Summary      Список заметок владельца
// @Description  Свежие сверху (по updated_at). ?user_id= должен совпасть с токеном.
// @Tags         notes
// @Produce      json
// @Param        user_id query string false "необязательно; если указан, должен совпадать с sub токена"
// @Param        limit   query int    false "размер страницы, по умолчанию 20, максимум 100"
// @Param        offset  query int    false "сколько записей пропустить, по умолчанию 0"
// @Success      200 {object} listResponse
// @Failure      400 {object} response.ErrorBody "некорректные limit или offset"
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Security     BearerAuth
// @Router       /notes [get]
func (t *Transport) list(c *gin.Context) {
	a, err := owner(c)
	if err != nil {
		abort(c, err)
		return
	}
	var q listQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		abort(c, response.Invalid("Некорректные limit и offset"))
		return
	}
	if q.Limit == 0 {
		q.Limit = defaultLimit
	}
	if q.Limit < 1 || q.Limit > maxLimit || q.Offset < 0 {
		abort(c, response.Invalid(fmt.Sprintf("limit от 1 до %d, offset от 0", maxLimit)))
		return
	}
	items, err := t.notes.List(c.Request.Context(), a, q.Limit, q.Offset)
	if err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 200, listResponse{Items: items, Limit: q.Limit, Offset: q.Offset})
}
