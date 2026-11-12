package http

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/middleware"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

// owner собирает, кто спрашивает: Subject ставит middleware.Auth из подписанного
// токена, Requested приезжает в ?user_id=. Заявка необязательна, сверку делает сервис.
func owner(c *gin.Context) (board.Access, error) {
	requested := strings.TrimSpace(c.Query("user_id"))
	if len(requested) > 64 {
		return board.Access{}, response.Invalid("Некорректный user_id")
	}
	return board.Access{Subject: middleware.UserID(c), Requested: requested}, nil
}

// boardID — id доски из пути.
func boardID(c *gin.Context) (string, error) {
	id := c.Param("id")
	if id == "" || len(id) > 64 {
		return "", response.Invalid("Некорректный id доски")
	}
	return id, nil
}

// abort переводит доменные ошибки в HTTP-ответы, остальное отдаёт response.Abort.
func abort(c *gin.Context, err error) {
	var conflict *board.ConflictError
	switch {
	case errors.As(err, &conflict):
		response.Abort(c, response.Problem(409, "version_conflict",
			fmt.Sprintf("Доска уже изменена, актуальная версия %d", conflict.Version)))
	case errors.Is(err, board.ErrNotFound):
		response.Abort(c, response.NotFound())
	case errors.Is(err, board.ErrAccess):
		response.Abort(c, response.Problem(403, "forbidden", board.ErrAccess.Error()))
	case errors.Is(err, board.ErrTitle), errors.Is(err, board.ErrScene):
		response.Abort(c, response.Invalid(err.Error()))
	default:
		response.Abort(c, err)
	}
}
