package onboarding

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ImportLegacy transfers the old snapshot once. Later starts never resurrect
// deleted data or overwrite progress now owned exclusively by Note.
func ImportLegacy(ctx context.Context, source, destination *pgxpool.Pool) error {
	tx, err := destination.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(271409)"); err != nil {
		return err
	}
	var done bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM onboarding.imports WHERE name='legacy-info-v1')").Scan(&done); err != nil {
		return err
	}
	if done {
		return tx.Commit(ctx)
	}
	// Repeatable-read snapshot prevents a partially inconsistent source export.
	src, err := source.Begin(ctx)
	if err != nil {
		return err
	}
	defer src.Rollback(ctx)
	if _, err = src.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ READ ONLY"); err != nil {
		return err
	}
	for _, table := range []string{"knowledge_documents", "checklist_tasks", "user_task_progress"} {
		var data []byte
		if err = src.QueryRow(ctx, "SELECT coalesce(json_agg(t),'[]'::json) FROM public."+table+" t").Scan(&data); err != nil {
			return fmt.Errorf("export %s: %w", table, err)
		}
		if _, err = tx.Exec(ctx, "INSERT INTO onboarding."+table+" SELECT * FROM json_populate_recordset(NULL::onboarding."+table+",$1::json) ON CONFLICT DO NOTHING", string(data)); err != nil {
			return fmt.Errorf("import %s: %w", table, err)
		}
	}
	if err = src.Commit(ctx); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO onboarding.imports(name) VALUES('legacy-info-v1')"); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
