package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

func (s *Service) Update(ctx context.Context, a board.Access, b board.Board) (board.Board, error) {
	userID, err := a.UserID()
	if err != nil {
		return board.Board{}, err
	}
	title, err := board.ValidateTitle(b.Title)
	if err != nil {
		return board.Board{}, err
	}
	scene, err := board.ValidateScene(b.Scene)
	if err != nil {
		return board.Board{}, err
	}
	b.Title, b.Scene, b.UserID = title, scene, userID
	return s.boards.Update(ctx, b)
}
