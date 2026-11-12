package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

func (s *Service) Get(ctx context.Context, a board.Access, id string) (board.Board, error) {
	userID, err := a.UserID()
	if err != nil {
		return board.Board{}, err
	}
	return s.boards.Get(ctx, id, userID)
}
