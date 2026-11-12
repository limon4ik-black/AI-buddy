package service

import (
	"context"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/idgen"
	"github.com/limon4ik-black/AI-buddy/note-service/internal/features/note"
)

func (s *Service) Create(ctx context.Context, a note.Access, content string) (note.Note, error) {
	userID, err := a.UserID()
	if err != nil {
		return note.Note{}, err
	}
	content, err = note.ValidateContent(content)
	if err != nil {
		return note.Note{}, err
	}
	return s.notes.Create(ctx, note.Note{ID: idgen.New(), UserID: userID, Content: content})
}
