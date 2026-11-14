package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Error struct {
	Status        int
	Code, Message string
}

func (e *Error) Error() string { return e.Message }

func Problem(status int, code, message string) error { return &Error{status, code, message} }

func Invalid(message string) error { return Problem(400, "invalid_request", message) }

func NotFound() error { return Problem(404, "not_found", "Объект не найден") }

func Handle(fn Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			var p *Error
			var pg *pgconn.PgError
			switch {
			case errors.As(err, &p):
			case errors.Is(err, pgx.ErrNoRows):
				p = &Error{404, "not_found", "Объект не найден"}
			case errors.As(err, &pg) && pg.Code == "23505":
				p = &Error{409, "conflict", "Запись уже существует"}
			default:
				slog.Error("request failed", "requestId", r.Context().Value(requestKey), "error", err)
				p = &Error{503, "service_unavailable", "Сервис временно недоступен. Повторите запрос"}
			}
			JSON(w, p.Status, map[string]any{"error": map[string]any{"code": p.Code, "message": p.Message, "requestId": r.Context().Value(requestKey)}})
		}
	}
}
