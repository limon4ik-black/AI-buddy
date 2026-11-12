package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

func (s *Service) Update(ctx context.Context, a note.Access, n note.Note) (note.Note, error) {
	userID, err := a.UserID()
	if err != nil {
		return note.Note{}, err
	}
	content, err := note.ValidateContent(n.Content)
	if err != nil {
		return note.Note{}, err
	}
	n.Content = content
	n.UserID = userID
	return s.notes.Update(ctx, n)
}
