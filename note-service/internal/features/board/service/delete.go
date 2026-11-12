package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

func (s *Service) Delete(ctx context.Context, a board.Access, id string) error {
	userID, err := a.UserID()
	if err != nil {
		return err
	}
	return s.boards.Delete(ctx, id, userID)
}
