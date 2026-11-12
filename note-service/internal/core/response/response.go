// Пакет response — единый формат ответов и ошибок.
package response

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Status        int
	Code, Message string
}

func (e *Error) Error() string { return e.Message }

// ErrorBody — форма тела ошибки, нужна только для документации: {"error":{"code":…,"message":…}}.
type ErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func Problem(status int, code, message string) error { return &Error{status, code, message} }
func Invalid(message string) error                   { return Problem(400, "invalid_request", message) }
func NotFound() error                                { return Problem(404, "not_found", "Объект не найден") }

// Abort отдаёт {"error":{code,message}}; неизвестную ошибку кладёт в контекст gin,
// чтобы её причину записал в лог middleware.Logger.
func Abort(c *gin.Context, err error) {
	var p *Error
	if !errors.As(err, &p) {
		_ = c.Error(err)
		p = &Error{500, "internal", "Внутренняя ошибка сервиса"}
	}
	c.AbortWithStatusJSON(p.Status, map[string]any{"error": map[string]any{"code": p.Code, "message": p.Message}})
}

func JSON(c *gin.Context, status int, value any) {
	if status == 204 {
		c.Status(status)
		return
	}
	c.JSON(status, value)
}

// maxBody — предел тела запроса по умолчанию; холсты просят свой, больший (см. board.MaxScene).
const maxBody = 1 << 20

// Decode перекладывает тело запроса в v. v — указатель на структуру запроса.
func Decode(c *gin.Context, v any) error { return DecodeLimit(c, v, maxBody) }

// DecodeLimit — то же, но с явным пределом тела. Обрезанное тело не пройдёт разбор,
// поэтому молчаливой потери данных не будет: клиент получит 400.
func DecodeLimit(c *gin.Context, v any, limit int64) error {
	if !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
		return Problem(415, "unsupported_media_type", "Ожидается application/json")
	}
	d := json.NewDecoder(io.LimitReader(c.Request.Body, limit))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return Invalid("Некорректное тело запроса")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return Invalid("Ожидается один JSON-объект")
	}
	return nil
}
