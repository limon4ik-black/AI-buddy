package employee

import (
	"context"
	"github.com/limon4ik-black/AI-buddy/shared/demodata"
)

// SeedDemo initializes employee profiles and the legacy onboarding source used
// only by Note's one-time importer. Employee exposes no onboarding runtime API;
// after import, Note owns all checklist/document/progress writes.
func (i *Service) SeedDemo(ctx context.Context) error {
	tx, err := i.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, u := range demodata.Interns() {
		if _, err = tx.Exec(ctx, "INSERT INTO profiles(user_id,full_name,team_id,position,start_date) VALUES($1,$2,$3,$4,$5::date) ON CONFLICT(user_id) DO NOTHING", u.ID, u.FullName, u.TeamID, u.Position, u.StartDate); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO people(id,full_name,position) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", u.ID, u.FullName, u.Position); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO team_members(team_id,person_id,role) SELECT team_id,user_id,'intern' FROM profiles WHERE user_id=$1 ON CONFLICT DO NOTHING", u.ID); err != nil {
			return err
		}
		for status, ids := range map[string][]string{"completed": u.Completed, "in_progress": u.InProgress} {
			for _, taskID := range ids {
				if _, err = tx.Exec(ctx, `INSERT INTO user_task_progress(user_id,task_id,status,completed_at)
     SELECT p.user_id,t.id,$3,CASE WHEN $3='completed' THEN now() ELSE NULL END
     FROM profiles p JOIN checklist_tasks t ON (t.team_id IS NULL OR t.team_id=p.team_id)
     WHERE p.user_id=$1 AND t.id=$2 ON CONFLICT DO NOTHING`, u.ID, taskID, status); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit(ctx)
}
