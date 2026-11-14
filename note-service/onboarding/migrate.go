package onboarding

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/note-service/migrations"
)

// Migrate is used by isolated tests; deployment uses the versioned SQL runner.
func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(271410)"); err != nil {
		return err
	}
	var exists bool
	if err = tx.QueryRow(ctx, "SELECT to_regclass('onboarding.checklist_tasks') IS NOT NULL").Scan(&exists); err != nil {
		return err
	}
	if !exists {
		if _, err = tx.Exec(ctx, string(migrations.Onboarding)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
