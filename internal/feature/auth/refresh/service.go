package refresh

import (
	"context"
	"errors"
	"fmt"

	"github.com/limon4ik-black/AI-buddy/internal/feature/auth"
)

var ErrInvalid = errors.New("invalid refresh token")

type Service struct {
	issuer  *auth.TokenIssuer
	refresh auth.RefreshStore
}

func NewService(issuer *auth.TokenIssuer, refresh auth.RefreshStore) *Service {
	return &Service{issuer: issuer, refresh: refresh}
}

type State struct {
	AccessToken  string
	ExpiresIn    int64
	RefreshToken string
	UserID       int64
}

func (s *Service) Refresh(ctx context.Context, token string) (*State, error) {
	userID, ok, err := s.refresh.Get(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("lookup refresh token: %w", err)
	}
	if !ok {
		return nil, ErrInvalid
	}

	access, expiresIn, err := s.issuer.IssueAccess(userID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	newToken, err := auth.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Ротация: новый сохраняем до удаления старого - при падении между
	// шагами юзер просто повторит запрос со старым токеном.
	if err := s.refresh.Save(ctx, newToken, userID); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}
	if err := s.refresh.Delete(ctx, token); err != nil {
		return nil, fmt.Errorf("delete old refresh token: %w", err)
	}

	return &State{
		AccessToken:  access,
		ExpiresIn:    expiresIn,
		RefreshToken: newToken,
		UserID:       userID,
	}, nil
}
