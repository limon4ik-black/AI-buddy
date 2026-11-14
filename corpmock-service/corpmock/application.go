package corpmock

import (
	"context"
	"net/http"

	"github.com/limon4ik-black/AI-buddy/shared/platform/config"
)

// NewApplication builds the mock without a database.
func NewApplication(ctx context.Context) (http.Handler, func(), error) {
	tokens, err := config.Secrets("CORP_SERVICE_TOKEN")
	if err != nil {
		return nil, nil, err
	}
	return New(tokens["CORP_SERVICE_TOKEN"]), func() {}, nil
}
