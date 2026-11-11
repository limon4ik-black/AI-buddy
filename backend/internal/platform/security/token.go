package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func TokenHash(v string) string { b := sha256.Sum256([]byte(v)); return hex.EncodeToString(b[:]) }

func Equal(a, b string) bool {
	return a != "" && b != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
