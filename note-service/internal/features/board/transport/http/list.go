package http

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// maxOverhead — запас к лимиту тела поверх сцены: название, version и скобки JSON.
const maxOverhead = 64 << 10

// listQuery — параметры GET /api/v1/boards.
type listQuery struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// listResponse — тело ответа GET /api/v1/boards. Сцен в списке нет: они тяжёлые,
// поэтому список отдаёт только метаданные, а сцена приезжает при открытии доски.
type listResponse struct {
	Items  []board.Board `json:"items"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// @Summary      Список досок владельца
// @Description  Свежие сверху (по updated_at), без сцен. ?user_id= должен совпасть с токеном.
// @Tags         boards
// @Produce      json
// @Param        user_id query string false "необязательно; если указан, должен совпадать с sub токена"
// @Param        limit   query int    false "размер страницы, по умолчанию 20, максимум 100"
// @Param        offset  query int    false "сколько записей пропустить, по умолчанию 0"
// @Success      200 {object} listResponse
// @Failure      400 {object} response.ErrorBody "некорректные limit или offset"
// @Failure      401 {object} response.ErrorBody "нет, битый или просроченный токен"
// @Failure      403 {object} response.ErrorBody "user_id не совпал с токеном"
// @Security     BearerAuth
// @Router       /boards [get]
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
	items, err := t.boards.List(c.Request.Context(), a, q.Limit, q.Offset)
	if err != nil {
		abort(c, err)
		return
	}
	response.JSON(c, 200, listResponse{Items: items, Limit: q.Limit, Offset: q.Offset})
}
