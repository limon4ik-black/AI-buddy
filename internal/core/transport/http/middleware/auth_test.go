package middleware

import (
	"crypto/ed25519"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func issueToken(t *testing.T, priv ed25519.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	s, err := tok.SignedString(priv)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAuthMiddleware(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	_, otherPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	valid := issueToken(t, priv, jwt.MapClaims{
		"user_id": float64(42),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	expired := issueToken(t, priv, jwt.MapClaims{
		"user_id": float64(42),
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})
	wrongKey := issueToken(t, otherPriv, jwt.MapClaims{"user_id": float64(42)})

	r := gin.New()
	r.GET("/test", AuthMiddleware(pub), func(c *gin.Context) {
		u, ok := FromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": u.ID})
	})

	cases := []struct {
		name   string
		header string
		want   int
		wantID int64
	}{
		{"no token", "", http.StatusUnauthorized, 0},
		{"not a bearer token", "Basic abc", http.StatusUnauthorized, 0},
		{"garbage token", "Bearer garbage", http.StatusUnauthorized, 0},
		{"expired token", "Bearer " + expired, http.StatusUnauthorized, 0},
		{"signed by another key", "Bearer " + wrongKey, http.StatusUnauthorized, 0},
		{"valid token", "Bearer " + valid, http.StatusOK, 42},
		{"case-insensitive scheme", "bearer " + valid, http.StatusOK, 42},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d", w.Code, tc.want)
			}
			if tc.want == http.StatusOK && w.Body.String() != `{"user_id":42}` {
				t.Fatalf("body = %s", w.Body.String())
			}
		})
	}
}
