package info

import (
	"net/http"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (i *Service) getChecklist(r *http.Request) (contracts.Checklist, error) {
	tasks, e := postgres.Query[contracts.Task](r.Context(), i.DB, `SELECT t.id,t.title,t.description,t.period,t.document_id,t.source_url,coalesce(p.status,'pending') AS status,p.completed_at
 FROM checklist_tasks t JOIN profiles u ON u.user_id=$1 LEFT JOIN user_task_progress p ON p.task_id=t.id AND p.user_id=u.user_id
 WHERE t.team_id IS NULL OR t.team_id=u.team_id ORDER BY t.sort_order,t.id`, httpapi.User(r).Subject)
	result := contracts.Checklist{Tasks: tasks, Total: len(tasks)}
	for _, t := range tasks {
		if t.Status == "completed" {
			result.Completed++
		} else if result.NextTask == nil {
			task := t
			result.NextTask = &task
		}
	}
	if result.Total > 0 {
		result.Percent = result.Completed * 100 / result.Total
	}
	return result, e
}

func (i *Service) checklist(w http.ResponseWriter, r *http.Request) error {
	v, e := i.getChecklist(r)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, v)
	return nil
}

func (i *Service) updateTask(w http.ResponseWriter, r *http.Request) error {
	v, e := httpapi.Decode[struct {
		Status string `json:"status"`
	}](r)
	if e != nil {
		return e
	}
	if v.Status != "pending" && v.Status != "in_progress" && v.Status != "completed" {
		return httpapi.Invalid("Неизвестный статус задачи")
	}
	result, e := i.DB.Exec(r.Context(), `INSERT INTO user_task_progress(user_id,task_id,status,completed_at)
 SELECT u.user_id,t.id,$3,CASE WHEN $3='completed' THEN now() ELSE NULL END FROM profiles u JOIN checklist_tasks t ON (t.team_id IS NULL OR t.team_id=u.team_id) WHERE u.user_id=$1 AND t.id=$2
 ON CONFLICT(user_id,task_id) DO UPDATE SET status=EXCLUDED.status,completed_at=CASE WHEN EXCLUDED.status='completed' THEN coalesce(user_task_progress.completed_at,EXCLUDED.completed_at) ELSE NULL END`, httpapi.User(r).Subject, r.PathValue("id"), v.Status)
	if e != nil {
		return e
	}
	if result.RowsAffected() == 0 {
		return httpapi.NotFound()
	}
	return i.checklist(w, r)
}
