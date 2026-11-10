package chat

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
	"github.com/limon4ik-black/AI-buddy/backend/migrations"
)

// NewApplication builds the service from environment configuration.
func NewApplication(ctx context.Context) (http.Handler, func(), error) {
	keys, err := config.JWT(false)
	if err != nil {
		return nil, nil, err
	}
	tokens, err := config.Secrets("GATEWAY_SERVICE_TOKEN", "CHAT_SERVICE_TOKEN", "RAG_SERVICE_TOKEN")
	if err != nil {
		return nil, nil, err
	}
	db, err := postgres.Open(ctx, os.Getenv("DATABASE_URL"), "chat", migrations.Chat)
	if err != nil {
		return nil, nil, err
	}
	svc := &Service{DB: db, Keys: keys, GatewayToken: tokens["GATEWAY_SERVICE_TOKEN"], ChatToken: tokens["CHAT_SERVICE_TOKEN"], RAGToken: tokens["RAG_SERVICE_TOKEN"], InfoURL: config.Env("INFO_URL", "http://info:8082"), RAGURL: config.Env("RAG_URL", "http://ragmock:8084"), Client: &http.Client{Timeout: 25 * time.Second}}
	return svc.Handler(), db.Close, nil
}
