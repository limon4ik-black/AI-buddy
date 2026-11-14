package security

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Keys struct {
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
	Now     func() time.Time
}

type Claims struct {
	Subject  string `json:"sub"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
	Expires  int64  `json:"exp"`
	Issued   int64  `json:"iat"`
}

func ParseKeys(seed, public string) (*Keys, error) {
	k := &Keys{Now: time.Now}
	if seed != "" {
		b, e := base64.StdEncoding.DecodeString(seed)
		if e != nil || len(b) != ed25519.SeedSize {
			return nil, errors.New("JWT_PRIVATE_SEED must be a base64 Ed25519 seed")
		}
		k.Private = ed25519.NewKeyFromSeed(b)
		k.Public = k.Private.Public().(ed25519.PublicKey)
	}
	if public != "" {
		b, e := base64.StdEncoding.DecodeString(public)
		if e != nil || len(b) != ed25519.PublicKeySize {
			return nil, errors.New("JWT_PUBLIC_KEY must be a base64 Ed25519 public key")
		}
		k.Public = b
	}
	if len(k.Public) == 0 {
		return nil, errors.New("JWT keys are required; run make configure")
	}
	return k, nil
}

func (k *Keys) Sign(subject, email string) string {
	now := k.Now()
	c := Claims{subject, email, "intern", "ai-buddy", "ai-buddy-api", now.Add(15 * time.Minute).Unix(), now.Unix()}
	b, _ := json.Marshal(c)
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA","typ":"JWT"}`)) + "." + base64.RawURLEncoding.EncodeToString(b)
	return payload + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(k.Private, []byte(payload)))
}

func (k *Keys) Verify(token string) (Claims, error) {
	var c Claims
	invalid := errors.New("invalid token")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c, invalid
	}
	h, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return c, invalid
	}
	var header struct{ Alg, Typ string }
	if json.Unmarshal(h, &header) != nil || header.Alg != "EdDSA" || header.Typ != "JWT" {
		return c, invalid
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[2])
	if e != nil || !ed25519.Verify(k.Public, []byte(parts[0]+"."+parts[1]), sig) {
		return c, invalid
	}
	b, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil || json.Unmarshal(b, &c) != nil {
		return c, invalid
	}
	if c.Subject == "" || c.Role != "intern" || c.Issuer != "ai-buddy" || c.Audience != "ai-buddy-api" || c.Expires <= k.Now().Unix() || c.Issued > k.Now().Add(time.Minute).Unix() {
		return c, invalid
	}
	return c, nil
}
