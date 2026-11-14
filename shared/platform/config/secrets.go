package config

import (
	"fmt"
	"os"

	"github.com/limon4ik-black/AI-buddy/shared/platform/security"
)

// Secrets validates the service credentials needed by one application.
func Secrets(names ...string) (map[string]string, error) {
	values := make(map[string]string, len(names))
	for _, name := range names {
		value := os.Getenv(name)
		if len(value) < 32 {
			return nil, fmt.Errorf("%s must contain at least 32 characters; run make configure", name)
		}
		values[name] = value
	}
	return values, nil
}

// JWT loads verification keys and, for Auth, the signing seed.
func JWT(signing bool) (*security.Keys, error) {
	seed := ""
	if signing {
		seed = os.Getenv("JWT_PRIVATE_SEED")
		if seed == "" {
			return nil, fmt.Errorf("JWT_PRIVATE_SEED is required by Auth")
		}
	}
	return security.ParseKeys(seed, os.Getenv("JWT_PUBLIC_KEY"))
}
