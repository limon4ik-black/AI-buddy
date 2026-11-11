package logout

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/internal/feature/auth"
)

type Service struct {
	refresh auth.RefreshStore
}

func NewService(refresh auth.RefreshStore) *Service {
	return &Service{refresh: refresh}
}

// Idempotent: токена уже нет - тоже успех.
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.refresh.Delete(ctx, token)
}
