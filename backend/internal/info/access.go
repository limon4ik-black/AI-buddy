package info

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpclient"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (i *Service) requests(w http.ResponseWriter, r *http.Request) error {
	v, e := postgres.Query[AccessRequest](r.Context(), i.DB, "SELECT * FROM access_requests WHERE user_id=$1 ORDER BY created_at DESC,id", httpapi.User(r).Subject)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, map[string]any{"items": v})
	return nil
}

func (i *Service) createRequest(w http.ResponseWriter, r *http.Request) error {
	v, e := httpapi.Decode[struct {
		System    string `json:"system"`
		Confirmed bool   `json:"confirmed"`
	}](r)
	if e != nil {
		return e
	}
	if !v.Confirmed {
		return httpapi.Invalid("Требуется явное подтверждение действия")
	}
	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 80 {
		return httpapi.Invalid("Нужен Idempotency-Key длиной от 16 до 80 символов")
	}
	var allowed bool
	e = i.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM corporate_tools t JOIN profiles p ON p.user_id=$1 WHERE t.id=$2 AND (t.team_id IS NULL OR t.team_id=p.team_id))", httpapi.User(r).Subject, v.System).Scan(&allowed)
	if e != nil {
		return e
	}
	if !allowed {
		return httpapi.Problem(403, "forbidden", "Инструмент недоступен")
	}
	tx, e := i.DB.Begin(r.Context())
	if e != nil {
		return e
	}
	defer tx.Rollback(r.Context())
	if _, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", httpapi.User(r).Subject+key); e != nil {
		return e
	}
	var raw []byte
	e = tx.QueryRow(r.Context(), "SELECT to_jsonb(a) FROM access_requests a WHERE user_id=$1 AND idempotency_key=$2", httpapi.User(r).Subject, key).Scan(&raw)
	if e == nil {
		var existing struct{ ID, System string }
		_ = json.Unmarshal(raw, &existing)
		if existing.System != v.System {
			return httpapi.Problem(409, "idempotency_conflict", "Ключ уже использован для другого инструмента")
		}
		_ = tx.Rollback(r.Context())
		return i.requestByID(w, r, existing.ID, 200)
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return e
	}
	// Stable key also makes a retry safe if the provider succeeded before a DB/network failure.
	result, e := httpclient.Call[contracts.CorpResult](r.Context(), i.Client, "POST", i.CorpURL+"/internal/v1/access-requests", i.CorpToken, "", contracts.CorpInput{UserID: httpapi.User(r).Subject, System: v.System, IdempotencyKey: key})
	if e != nil {
		return httpapi.Problem(502, "corporate_unavailable", "Не удалось отправить заявку. Повторите с тем же ключом")
	}
	if result.ID == "" || result.Status != "submitted" {
		return httpapi.Problem(502, "invalid_corporate_response", "Корпоративный сервис вернул некорректный результат")
	}
	id := idgen.New()
	_, e = tx.Exec(r.Context(), "INSERT INTO access_requests(id,user_id,system,status,external_id,idempotency_key) VALUES($1,$2,$3,$4,$5,$6)", id, httpapi.User(r).Subject, v.System, result.Status, result.ID, key)
	if e != nil {
		return e
	}
	if e = tx.Commit(r.Context()); e != nil {
		return e
	}
	return i.requestByID(w, r, id, 201)
}

func (i *Service) requestByID(w http.ResponseWriter, r *http.Request, id string, status int) error {
	v, e := postgres.One[AccessRequest](r.Context(), i.DB, "SELECT * FROM access_requests WHERE id=$1 AND user_id=$2", id, httpapi.User(r).Subject)
	if e != nil {
		return e
	}
	httpapi.JSON(w, status, v)
	return nil
}
