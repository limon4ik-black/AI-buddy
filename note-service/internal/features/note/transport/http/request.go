package http

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/middleware"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/response"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

// owner собирает, кто спрашивает: Subject ставит middleware.Auth из подписанного
// токена, Requested приезжает в ?user_id=. Заявка необязательна, сверку делает сервис.
func owner(c *gin.Context) (note.Access, error) {
	requested := strings.TrimSpace(c.Query("user_id"))
	if len(requested) > 64 {
		return note.Access{}, response.Invalid("Некорректный user_id")
	}
	return note.Access{Subject: middleware.UserID(c), Requested: requested}, nil
}

// noteID — id заметки из пути.
func noteID(c *gin.Context) (string, error) {
	id := c.Param("id")
	if id == "" || len(id) > 64 {
		return "", response.Invalid("Некорректный id заметки")
	}
	return id, nil
}

// abort переводит доменные ошибки в HTTP-ответы, остальное отдаёт response.Abort.
func abort(c *gin.Context, err error) {
	var conflict *note.ConflictError
	switch {
	case errors.As(err, &conflict):
		response.Abort(c, response.Problem(409, "version_conflict",
			fmt.Sprintf("Заметка уже изменена, актуальная версия %d", conflict.Version)))
	case errors.Is(err, note.ErrNotFound):
		response.Abort(c, response.NotFound())
	case errors.Is(err, note.ErrAccess):
		response.Abort(c, response.Problem(403, "forbidden", note.ErrAccess.Error()))
	case errors.Is(err, note.ErrContent):
		response.Abort(c, response.Invalid(note.ErrContent.Error()))
	default:
		response.Abort(c, err)
	}
}
