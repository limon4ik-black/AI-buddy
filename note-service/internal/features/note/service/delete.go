package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

func (s *Service) Delete(ctx context.Context, a note.Access, id string) error {
	userID, err := a.UserID()
	if err != nil {
		return err
	}
	return s.notes.Delete(ctx, id, userID)
}
