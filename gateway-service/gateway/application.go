package gateway

import (
	"context"
	"net/http"

	"github.com/limon4ik-black/AI-buddy/shared/platform/config"
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
	handler, err := New(Config{Keys: keys, Token: tokens["GATEWAY_SERVICE_TOKEN"], Origin: config.Env("APP_ORIGIN", "http://localhost:8080"), AuthURL: config.Env("AUTH_URL", "http://auth:8081"), EmployeeURL: config.Env("EMPLOYEE_URL", "http://employee:8082"), NoteURL: config.Env("NOTE_URL", "http://note_service:8079"), ChatURL: config.Env("CHAT_URL", "http://chat:8083"), ChatWSURL: config.Env("CHAT_WS_URL", "http://chat:8083")})
	return handler, func() {}, err
}
