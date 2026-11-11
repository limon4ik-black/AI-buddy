package login

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/limon4ik-black/AI-buddy/internal/feature/auth"
)

type GoogleProfile struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Service struct {
	conf    *oauth2.Config
	users   auth.UserStore
	issuer  *auth.TokenIssuer
	refresh auth.RefreshStore
}

func NewService(conf *oauth2.Config, users auth.UserStore, issuer *auth.TokenIssuer, refresh auth.RefreshStore) *Service {
	return &Service{conf: conf, users: users, issuer: issuer, refresh: refresh}
}

type LoginState struct {
	AccessToken  string
	ExpiresIn    int64
	RefreshToken string
	UserID       int64
	Email        string
	Name         string
}

func (s *Service) Login(ctx context.Context, code string) (*LoginState, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tok, err := s.conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	profile, err := s.fetchProfile(ctx, tok)
	if err != nil {
		return nil, fmt.Errorf("fetch profile: %w", err)
	}

	userID := s.users.FindOrCreate("google", profile.Sub, profile.Email, profile.Name)

	access, expiresIn, err := s.issuer.IssueAccess(userID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}
	refreshToken, err := auth.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	if err := s.refresh.Save(ctx, refreshToken, userID); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &LoginState{
		AccessToken:  access,
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		UserID:       userID,
		Email:        profile.Email,
		Name:         profile.Name,
	}, nil
}

func (s *Service) fetchProfile(ctx context.Context, tok *oauth2.Token) (*GoogleProfile, error) {
	client := s.conf.Client(ctx, tok)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var p GoogleProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}
