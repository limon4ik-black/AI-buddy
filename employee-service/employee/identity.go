package employee

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/limon4ik-black/AI-buddy/shared/demodata"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
)

// Auth's signed subject is the only account ID. Demo profiles created before an
// existing Auth email was seeded may live under a template ID. Copy those details
// to the signed account, preserving customized profiles. Keep
// the original row and its access requests: directory IDs alone do not prove that
// historical requests belong to this account, so they must not be reassigned.
func (i *Service) ensureProfile(r *http.Request) error {
	claims := httpapi.User(r)
	ctx := r.Context()
	tx, err := i.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 1))", claims.Subject); err != nil {
		return err
	}
	for _, intern := range demodata.Interns() {
		if !strings.EqualFold(claims.Email, intern.Email) || intern.ID == claims.Subject {
			continue
		}
		var copied string
		err = tx.QueryRow(ctx, `INSERT INTO profiles(user_id,full_name,team_id,position,start_date)
 SELECT $1,full_name,team_id,position,start_date FROM profiles WHERE user_id=$2
 ON CONFLICT(user_id) DO UPDATE SET full_name=EXCLUDED.full_name,team_id=EXCLUDED.team_id,
 position=EXCLUDED.position,start_date=EXCLUDED.start_date
 WHERE profiles.full_name='Стажёр' AND profiles.position='Backend-стажёр' AND profiles.team_id='backend'
 RETURNING user_id`, claims.Subject, intern.ID).Scan(&copied)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		break
	}
	if _, err = tx.Exec(ctx, "INSERT INTO profiles(user_id) VALUES($1) ON CONFLICT DO NOTHING", claims.Subject); err != nil {
		return err
	}
	if err = syncProfileDirectory(ctx, tx, claims.Subject); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func syncProfileDirectory(ctx context.Context, tx pgx.Tx, userID string) error {
	if _, err := tx.Exec(ctx, `INSERT INTO people(id,full_name,position)
 SELECT user_id,full_name,position FROM profiles WHERE user_id=$1
 ON CONFLICT(id) DO UPDATE SET full_name=EXCLUDED.full_name,position=EXCLUDED.position`, userID); err != nil {
		return err
	}
	// A repaired default profile may change teams. Remove only stale trainee
	// membership; buddy/lead/engineer assignments are independent directory data.
	if _, err := tx.Exec(ctx, `DELETE FROM team_members
 WHERE person_id=$1 AND role='intern'
 AND team_id<>(SELECT team_id FROM profiles WHERE user_id=$1)`, userID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO team_members(team_id,person_id,role)
 SELECT team_id,user_id,'intern' FROM profiles WHERE user_id=$1
 ON CONFLICT(team_id,person_id) DO NOTHING`, userID)
	return err
}
