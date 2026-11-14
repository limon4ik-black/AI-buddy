package onboarding

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/demodata"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpclient"
	"github.com/limon4ik-black/AI-buddy/shared/platform/postgres"
	"github.com/limon4ik-black/AI-buddy/shared/platform/security"
)

type Service struct {
	DB                                                 *pgxpool.Pool
	Keys                                               *security.Keys
	GatewayToken, ChatToken, IndexerToken, EmployeeURL string
	Client                                             *http.Client
}
type contextKey struct{}

func (s *Service) Handler() http.Handler {
	m := http.NewServeMux()
	public := func(pattern string, fn httpapi.Endpoint) {
		m.HandleFunc(pattern, httpapi.Handle(httpapi.ServiceToken(s.GatewayToken, httpapi.Authenticated(s.Keys, s.withEmployee(fn)))))
	}
	public("GET /api/v1/checklist", s.checklist)
	public("PATCH /api/v1/checklist/tasks/{id}", s.updateTask)
	public("GET /api/v1/information", s.documents)
	public("GET /api/v1/information/{id}", s.document)
	m.HandleFunc("GET /internal/v1/context", httpapi.Handle(httpapi.ServiceToken(s.ChatToken, httpapi.Authenticated(s.Keys, s.withEmployee(s.userContext)))))
	m.HandleFunc("GET /internal/v1/users/{userId}/context", httpapi.Handle(httpapi.ServiceToken(s.ChatToken, httpapi.Authenticated(s.Keys, func(w http.ResponseWriter, r *http.Request) error {
		if r.PathValue("userId") != httpapi.User(r).Subject {
			return httpapi.Problem(403, "identity_mismatch", "Нельзя запрашивать контекст другого пользователя")
		}
		return s.withEmployee(s.userContext)(w, r)
	}))))
	m.HandleFunc("GET /internal/v1/knowledge/documents", httpapi.Handle(httpapi.ServiceToken(s.IndexerToken, s.export)))
	m.HandleFunc("GET /internal/v1/checklist/templates", httpapi.Handle(httpapi.ServiceToken(s.IndexerToken, s.templates)))
	// Directory data stays owned by Employee; RAG/indexers see one Note API.
	if u, err := url.Parse(s.EmployeeURL); err == nil && u.Host != "" {
		proxy := httputil.NewSingleHostReverseProxy(u)
		for _, path := range []string{"GET /internal/v1/teams", "GET /internal/v1/teams/{id}", "GET /internal/v1/services", "GET /internal/v1/services/{id}"} {
			m.HandleFunc(path, httpapi.Handle(httpapi.ServiceToken(s.IndexerToken, func(w http.ResponseWriter, r *http.Request) error { proxy.ServeHTTP(w, r); return nil })))
		}
	}
	return httpapi.Middleware(m)
}

func (s *Service) withEmployee(fn httpapi.Endpoint) httpapi.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := httpapi.User(r).Subject
		ctx, err := httpclient.Call[contracts.UserContext](r.Context(), s.Client, "GET", s.EmployeeURL+"/internal/v1/users/"+url.PathEscape(id)+"/context", s.ChatToken, r.Header.Get("Authorization"), nil)
		if err != nil {
			return httpapi.Problem(502, "employee_unavailable", "Не удалось получить профиль Employee")
		}
		if ctx.Profile.UserID != id {
			return httpapi.Problem(502, "identity_mismatch", "Employee вернул чужой профиль")
		}
		if err = s.linkProgress(r); err != nil {
			return err
		}
		return fn(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, ctx)))
	}
}
func employeeContext(r *http.Request) contracts.UserContext {
	return r.Context().Value(contextKey{}).(contracts.UserContext)
}

func (s *Service) linkProgress(r *http.Request) error {
	claims := httpapi.User(r)
	source := claims.Subject
	for _, u := range demodata.Interns() {
		if strings.EqualFold(u.Email, claims.Email) {
			source = u.ID
			break
		}
	}
	tx, err := s.DB.Begin(r.Context())
	if err != nil {
		return err
	}
	defer tx.Rollback(r.Context())
	if _, err = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtextextended($1,2))", claims.Subject); err != nil {
		return err
	}
	tag, err := tx.Exec(r.Context(), "INSERT INTO onboarding.identity_links(user_id,source_user_id) VALUES($1,$2) ON CONFLICT DO NOTHING", claims.Subject, source)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 && source != claims.Subject {
		if _, err = tx.Exec(r.Context(), `INSERT INTO onboarding.user_task_progress SELECT $1,task_id,status,completed_at FROM onboarding.user_task_progress WHERE user_id=$2 ON CONFLICT DO NOTHING`, claims.Subject, source); err != nil {
			return err
		}
	}
	return tx.Commit(r.Context())
}

func (s *Service) getChecklist(r *http.Request) (contracts.Checklist, error) {
	ctx := employeeContext(r)
	tasks, err := postgres.Query[contracts.Task](r.Context(), s.DB, `SELECT t.id,t.title,t.description,t.period,t.document_id,t.source_url,coalesce(p.status,'pending') AS status,p.completed_at
 FROM onboarding.checklist_tasks t LEFT JOIN onboarding.user_task_progress p ON p.task_id=t.id AND p.user_id=$1
 WHERE t.team_id IS NULL OR t.team_id=$2 ORDER BY t.sort_order,t.id`, ctx.Profile.UserID, ctx.Profile.TeamID)
	result := contracts.Checklist{Tasks: tasks, Total: len(tasks)}
	for _, task := range tasks {
		if task.Status == "completed" {
			result.Completed++
		} else if result.NextTask == nil {
			v := task
			result.NextTask = &v
		}
	}
	if result.Total > 0 {
		result.Percent = result.Completed * 100 / result.Total
	}
	return result, err
}
func (s *Service) checklist(w http.ResponseWriter, r *http.Request) error {
	v, err := s.getChecklist(r)
	if err != nil {
		return err
	}
	httpapi.JSON(w, 200, v)
	return nil
}
func (s *Service) updateTask(w http.ResponseWriter, r *http.Request) error {
	v, err := httpapi.Decode[struct {
		Status string `json:"status"`
	}](r)
	if err != nil {
		return err
	}
	if v.Status != "pending" && v.Status != "in_progress" && v.Status != "completed" {
		return httpapi.Invalid("Неизвестный статус задачи")
	}
	ctx := employeeContext(r)
	tag, err := s.DB.Exec(r.Context(), `INSERT INTO onboarding.user_task_progress(user_id,task_id,status,completed_at)
 SELECT $1,id,$3,CASE WHEN $3='completed' THEN now() ELSE NULL END FROM onboarding.checklist_tasks WHERE id=$2 AND (team_id IS NULL OR team_id=$4)
 ON CONFLICT(user_id,task_id) DO UPDATE SET status=EXCLUDED.status,completed_at=CASE WHEN EXCLUDED.status='completed' THEN coalesce(user_task_progress.completed_at,EXCLUDED.completed_at) ELSE NULL END`, ctx.Profile.UserID, r.PathValue("id"), v.Status, ctx.Profile.TeamID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpapi.NotFound()
	}
	return s.checklist(w, r)
}
func (s *Service) getDocuments(r *http.Request) ([]contracts.Document, error) {
	return postgres.Query[contracts.Document](r.Context(), s.DB, `SELECT * FROM onboarding.knowledge_documents WHERE deleted_at IS NULL AND (visibility='company' OR team_id=$1) ORDER BY id`, employeeContext(r).Profile.TeamID)
}
func (s *Service) documents(w http.ResponseWriter, r *http.Request) error {
	v, err := s.getDocuments(r)
	if err != nil {
		return err
	}
	httpapi.JSON(w, 200, map[string]any{"items": v})
	return nil
}
func (s *Service) document(w http.ResponseWriter, r *http.Request) error {
	v, err := postgres.One[contracts.Document](r.Context(), s.DB, `SELECT * FROM onboarding.knowledge_documents WHERE id=$1 AND deleted_at IS NULL AND (visibility='company' OR team_id=$2)`, r.PathValue("id"), employeeContext(r).Profile.TeamID)
	if err != nil {
		return err
	}
	httpapi.JSON(w, 200, v)
	return nil
}
func (s *Service) userContext(w http.ResponseWriter, r *http.Request) error {
	ctx := employeeContext(r)
	var err error
	ctx.Onboarding, err = s.getChecklist(r)
	if err != nil {
		return err
	}
	docs, err := s.getDocuments(r)
	if err != nil {
		return err
	}
	ctx.AllowedDocumentIDs = []string{}
	for _, doc := range docs {
		ctx.AllowedDocumentIDs = append(ctx.AllowedDocumentIDs, doc.ID)
	}
	// Company-visible directory references remain valid citation identifiers.
	teams, err := httpclient.Call[struct {
		Items []contracts.Team `json:"items"`
	}](r.Context(), s.Client, "GET", s.EmployeeURL+"/internal/v1/teams", s.IndexerToken, "", nil)
	if err != nil {
		return httpapi.Problem(502, "employee_unavailable", "Не удалось получить команды")
	}
	for _, team := range teams.Items {
		ctx.AllowedDocumentIDs = append(ctx.AllowedDocumentIDs, "team:"+team.ID)
	}
	services, err := httpclient.Call[struct {
		Items []contracts.TechnicalService `json:"items"`
	}](r.Context(), s.Client, "GET", s.EmployeeURL+"/internal/v1/services", s.IndexerToken, "", nil)
	if err != nil {
		return httpapi.Problem(502, "employee_unavailable", "Не удалось получить сервисы")
	}
	for _, service := range services.Items {
		ctx.AllowedDocumentIDs = append(ctx.AllowedDocumentIDs, "service:"+service.ID)
	}
	httpapi.JSON(w, 200, ctx)
	return nil
}
func (s *Service) export(w http.ResponseWriter, r *http.Request) error {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 100 {
			return httpapi.Invalid("limit должен быть от 1 до 100")
		}
		limit = v
	}
	docs, err := postgres.Query[contracts.Document](r.Context(), s.DB, "SELECT * FROM onboarding.knowledge_documents WHERE id>$1 ORDER BY id LIMIT $2", r.URL.Query().Get("after"), limit+1)
	if err != nil {
		return err
	}
	next := ""
	if len(docs) > limit {
		docs = docs[:limit]
		next = docs[len(docs)-1].ID
	}
	httpapi.JSON(w, 200, map[string]any{"documents": docs, "nextCursor": next, "syncMode": "full_snapshot", "aclRequired": true})
	return nil
}

type TaskTemplate struct {
	ID          string  `json:"id"`
	TeamID      *string `json:"teamId" db:"team_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Period      string  `json:"period"`
	SortOrder   int     `json:"sortOrder" db:"sort_order"`
	DocumentID  *string `json:"documentId" db:"document_id"`
	SourceURL   string  `json:"sourceUrl" db:"source_url"`
}

func (s *Service) templates(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("teamId")
	if id == "" {
		return httpapi.Invalid("Нужен teamId")
	}
	// Validate a real team in Employee instead of accepting arbitrary client data.
	u, err := http.NewRequestWithContext(r.Context(), "GET", s.EmployeeURL+"/internal/v1/teams/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	u.Header.Set("X-Service-Token", s.IndexerToken)
	res, err := s.Client.Do(u)
	if err != nil {
		return httpapi.Problem(502, "employee_unavailable", "Employee недоступен")
	}
	res.Body.Close()
	if res.StatusCode == 404 {
		return httpapi.NotFound()
	}
	if res.StatusCode != 200 {
		return httpapi.Problem(502, "employee_unavailable", "Employee недоступен")
	}
	v, err := postgres.Query[TaskTemplate](r.Context(), s.DB, "SELECT * FROM onboarding.checklist_tasks WHERE team_id IS NULL OR team_id=$1 ORDER BY sort_order,id", id)
	if err != nil {
		return err
	}
	httpapi.JSON(w, 200, map[string]any{"items": v, "teamId": id, "synthetic": true})
	return nil
}
