package auth

import (
	"context"
	"net/http"
	"os"

	"github.com/limon4ik-black/AI-buddy/shared/migrations"
	"github.com/limon4ik-black/AI-buddy/shared/platform/config"
	"github.com/limon4ik-black/AI-buddy/shared/platform/postgres"
)

// NewApplication builds the service from environment configuration.
func NewApplication(ctx context.Context) (http.Handler, func(), error) {
	keys, err := config.JWT(true)
	if err != nil {
		return nil, nil, err
	}
	db, err := postgres.Open(ctx, os.Getenv("DATABASE_URL"), "auth", migrations.Auth, migrations.AuthGoogle, migrations.AuthEmail, migrations.AuthPending, migrations.AuthPassword)
	if err != nil {
		return nil, nil, err
	}
	svc := &Service{DB: db, Keys: keys, SecureCookie: config.Env("COOKIE_SECURE", "false") == "true", Google: googleConf(), Mailer: newMailer()}
	if config.Env("SEED_DEMO", "false") == "true" {
		if err = svc.SeedDemo(ctx); err != nil {
			db.Close()
			return nil, nil, err
		}
	}
	return svc.Handler(), db.Close, nil
}
