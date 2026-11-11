package gateway

import (
	"context"
	"net/http"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
)

// NewApplication builds the gateway without a database.
func NewApplication(ctx context.Context) (http.Handler, func(), error) {
	keys, err := config.JWT(false)
	if err != nil {
		return nil, nil, err
	}
	tokens, err := config.Secrets("GATEWAY_SERVICE_TOKEN")
	if err != nil {
		return nil, nil, err
	}
	handler, err := New(Config{Keys: keys, Token: tokens["GATEWAY_SERVICE_TOKEN"], Origin: config.Env("APP_ORIGIN", "http://localhost:8080"), AuthURL: config.Env("AUTH_URL", "http://auth:8081"), InfoURL: config.Env("INFO_URL", "http://info:8082"), ChatURL: config.Env("CHAT_URL", "http://chat:8083")})
	return handler, func() {}, err
}
