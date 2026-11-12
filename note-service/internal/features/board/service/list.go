package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

func (s *Service) List(ctx context.Context, a board.Access, limit, offset int) ([]board.Board, error) {
	userID, err := a.UserID()
	if err != nil {
		return nil, err
	}
	return s.boards.List(ctx, userID, limit, offset)
}
