package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

type TokenIssuer struct {
	priv ed25519.PrivateKey
}

func NewTokenIssuer(priv ed25519.PrivateKey) *TokenIssuer {
	return &TokenIssuer{priv: priv}
}

func (i *TokenIssuer) IssueAccess(userID int64) (token string, expiresIn int64, err error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     now.Unix(),
		"exp":     now.Add(AccessTokenTTL).Unix(),
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(i.priv)
	if err != nil {
		return "", 0, err
	}
	return token, int64(AccessTokenTTL / time.Second), nil
}

// Refresh-токен - не JWT и не подписывается: его ценность в записи в хранилище,
// поэтому он отзываем в любой момент.
func NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
