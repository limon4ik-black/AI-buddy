package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

func (s *Service) List(ctx context.Context, a note.Access, limit, offset int) ([]note.Note, error) {
	userID, err := a.UserID()
	if err != nil {
		return nil, err
	}
	return s.notes.List(ctx, userID, limit, offset)
}
