package employee

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/limon4ik-black/AI-buddy/shared/migrations"
	"github.com/limon4ik-black/AI-buddy/shared/platform/config"
	"github.com/limon4ik-black/AI-buddy/shared/platform/postgres"
)

// NewApplication builds the service from environment configuration.
func NewApplication(ctx context.Context) (http.Handler, func(), error) {
	keys, err := config.JWT(false)
	if err != nil {
		return nil, nil, err
	}
	tokens, err := config.Secrets("GATEWAY_SERVICE_TOKEN", "CHAT_SERVICE_TOKEN", "INDEXER_SERVICE_TOKEN", "CORP_SERVICE_TOKEN")
	if err != nil {
		return nil, nil, err
	}
	// Retain the existing migration identity and database: changing "info" here
	// would replay historical CREATE TABLE migrations against preserved volumes.
	db, err := postgres.Open(ctx, os.Getenv("DATABASE_URL"), "info", migrations.Info, migrations.InfoMVP)
	if err != nil {
		return nil, nil, err
	}
	svc := &Service{DB: db, Keys: keys, GatewayToken: tokens["GATEWAY_SERVICE_TOKEN"], ChatToken: tokens["CHAT_SERVICE_TOKEN"], IndexerToken: tokens["INDEXER_SERVICE_TOKEN"], CorpToken: tokens["CORP_SERVICE_TOKEN"], CorpURL: config.Env("CORP_URL", "http://corpmock:8085"), Client: &http.Client{Timeout: 25 * time.Second}}
	if config.Env("SEED_DEMO", "false") == "true" {
		if err = svc.SeedDemo(ctx); err != nil {
			db.Close()
			return nil, nil, err
		}
	}
	return svc.Handler(), db.Close, nil
}
