package ragmock

import (
	"context"
	"net/http"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
)

// NewApplication builds the mock without a database.
func NewApplication(ctx context.Context) (http.Handler, func(), error) {
	tokens, err := config.Secrets("RAG_SERVICE_TOKEN")
	if err != nil {
		return nil, nil, err
	}
	return New(tokens["RAG_SERVICE_TOKEN"]), func() {}, nil
}
