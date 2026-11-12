package chat

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("chat not found")
	ErrMessage  = errors.New("message must contain 1 to 8000 characters")
	ErrAI       = errors.New("ai unavailable")
)

type Created struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}
type Summary struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updated_at"`
}
type Message struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
type Turn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Repository interface {
	Create(context.Context, uuid.UUID) (Created, error)
	List(context.Context, uuid.UUID) ([]Summary, error)
	History(context.Context, uuid.UUID, uuid.UUID) ([]Message, error)
	Delete(context.Context, uuid.UUID, uuid.UUID) error
	Append(context.Context, uuid.UUID, uuid.UUID, string, string) (Message, error)
}
type AI interface {
	Answer(context.Context, string, []Turn) (string, error)
}
type Service struct {
	Repository
	AI AI
}

func (s *Service) Send(ctx context.Context, user, id uuid.UUID, content string) (string, error) {
	content = strings.TrimSpace(content)
	if !utf8.ValidString(content) || content == "" || utf8.RuneCountInString(content) > 8000 {
		return "", ErrMessage
	}
	current, err := s.Append(ctx, user, id, "user", content)
	if err != nil {
		return "", err
	}
	messages, err := s.History(ctx, user, id)
	if err != nil {
		return "", err
	}
	history := make([]Turn, 0)
	for _, m := range messages {
		// Only messages preceding this request belong to its context.
		if m.ID < current.ID {
			history = append(history, Turn{m.Role, m.Content})
		}
	}
	if len(history) > 20 {
		history = history[len(history)-20:]
	}
	answer, err := s.AI.Answer(ctx, content, history)
	if err != nil {
		return "", ErrAI
	}
	if _, err = s.Append(ctx, user, id, "assistant", answer); err != nil {
		return "", err
	}
	return answer, nil
}
