package service

import (
	"context"
	"encoding/json"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/idgen"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/board"
)

func (s *Service) Create(ctx context.Context, a board.Access, title string, scene json.RawMessage) (board.Board, error) {
	userID, err := a.UserID()
	if err != nil {
		return board.Board{}, err
	}
	title, err = board.ValidateTitle(title)
	if err != nil {
		return board.Board{}, err
	}
	scene, err = board.ValidateScene(scene)
	if err != nil {
		return board.Board{}, err
	}
	return s.boards.Create(ctx, board.Board{ID: idgen.New(), UserID: userID, Title: title, Scene: scene})
}
