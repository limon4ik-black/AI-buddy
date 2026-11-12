package board

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxTitle и MaxScene — пределы полей; держать в согласии с CHECK в board.boards.
const (
	MaxTitle = 200
	MaxScene = 8 << 20
)

var (
	ErrTitle    = errors.New("Название доски должно быть не длиннее 200 символов")
	ErrScene    = errors.New("Сцена должна быть JSON-объектом не больше 8 МиБ")
	ErrNotFound = errors.New("board not found")
	ErrAccess   = errors.New("Нельзя обращаться к доскам другого пользователя")
)

// Access — кто спрашивает: Subject из подписанного токена и Requested из запроса.
type Access struct {
	Subject   string
	Requested string
}

// UserID сверяет заявленного владельца с токеном и возвращает доверенный.
// Заявка необязательна: браузерный клиент шлёт только токен.
func (a Access) UserID() (string, error) {
	if a.Subject == "" || (a.Requested != "" && a.Requested != a.Subject) {
		return "", ErrAccess
	}
	return a.Subject, nil
}

// ConflictError — правку отклонила оптимистичная блокировка.
type ConflictError struct{ Version int }

func (e *ConflictError) Error() string {
	return fmt.Sprintf("version conflict, current version %d", e.Version)
}

// Board — доска: название и сцена Excalidraw как есть, без разбора на элементы.
// В списке Scene пустая: сцены тяжёлые, list их не выбирает, а omitempty не отдаёт.
type Board struct {
	ID        string          `json:"id"`
	UserID    string          `json:"userId" db:"user_id"`
	Title     string          `json:"title"`
	Scene     json.RawMessage `json:"scene,omitempty" swaggertype:"object"` // swag не умеет RawMessage без подсказки
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time       `json:"updatedAt" db:"updated_at"`
}

// ValidateTitle обрезает пробелы и проверяет границы названия.
func ValidateTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if utf8.RuneCountInString(title) > MaxTitle {
		return "", ErrTitle
	}
	return title, nil
}

// ValidateScene проверяет размер и что приехал именно JSON-объект. Внутренний формат
// сцены не разбираем: он принадлежит Excalidraw и может поменяться, нам нужна форма.
func ValidateScene(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 || len(raw) > MaxScene || !json.Valid(raw) {
		return nil, ErrScene
	}
	if trimmed := bytes.TrimSpace(raw); len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, ErrScene
	}
	return raw, nil
}
