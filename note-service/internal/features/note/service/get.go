package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

func (s *Service) Get(ctx context.Context, a note.Access, id string) (note.Note, error) {
	userID, err := a.UserID()
	if err != nil {
		return note.Note{}, err
	}
	return s.notes.Get(ctx, id, userID)
}
