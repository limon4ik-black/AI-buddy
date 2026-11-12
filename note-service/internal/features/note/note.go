package note

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxContent — предел текста заметки; держать в согласии с CHECK в note.notes.
const MaxContent = 20000

// ErrContent — текст вне границ.
var ErrContent = errors.New("Текст заметки должен содержать от 1 до 20000 символов")

// ErrNotFound — заметки нет или она принадлежит другому.
var ErrNotFound = errors.New("note not found")

// ErrAccess — заявленный владелец не совпал с тем, что в токене.
var ErrAccess = errors.New("Нельзя обращаться к заметкам другого пользователя")

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

type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId" db:"user_id"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// ValidateContent обрезает пробелы и проверяет границы текста.
func ValidateContent(raw string) (string, error) {
	content := strings.TrimSpace(raw)
	if n := utf8.RuneCountInString(content); n < 1 || n > MaxContent {
		return "", ErrContent
	}
	return content, nil
}
