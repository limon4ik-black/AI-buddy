package chat

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpclient"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
)

func (c *Service) answer(w http.ResponseWriter, r *http.Request) error {
	if _, e := c.owned(r); e != nil {
		return e
	}
	tx, e := c.DB.Begin(r.Context())
	if e != nil {
		return e
	}
	defer tx.Rollback(r.Context())
	if _, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", r.PathValue("id")); e != nil {
		return e
	}
	var msg Message
	rows, e := tx.Query(r.Context(), "SELECT * FROM messages WHERE id=$1 AND chat_id=$2 AND role='user'", r.PathValue("messageId"), r.PathValue("id"))
	if e != nil {
		return e
	}
	msg, e = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Message])
	if e != nil {
		return e
	}
	existing, e := tx.Query(r.Context(), "SELECT * FROM messages WHERE reply_to=$1", msg.ID)
	if e != nil {
		return e
	}
	previous, e := pgx.CollectExactlyOneRow(existing, pgx.RowToStructByName[Message])
	if e == nil {
		httpapi.JSON(w, 200, previous)
		return nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return e
	}
	ctx, e := httpclient.Call[contracts.UserContext](r.Context(), c.Client, "GET", c.InfoURL+"/internal/v1/context", c.ChatToken, r.Header.Get("Authorization"), nil)
	if e != nil {
		return httpapi.Problem(502, "context_unavailable", "Не удалось получить актуальный контекст. Сообщение сохранено, повторите генерацию")
	}
	historyRows, e := tx.Query(r.Context(), "SELECT role,content FROM (SELECT role,content,sequence FROM messages WHERE chat_id=$1 AND sequence<$2 ORDER BY sequence DESC LIMIT 20) h ORDER BY sequence", msg.ChatID, msg.Sequence)
	if e != nil {
		return e
	}
	history, e := pgx.CollectRows(historyRows, pgx.RowToStructByName[contracts.Turn])
	if e != nil {
		return e
	}
	answer, e := httpclient.Call[contracts.GenerateOutput](r.Context(), c.Client, "POST", c.RAGURL+"/internal/v1/generate", c.RAGToken, "", contracts.GenerateInput{RequestID: httpapi.RequestID(r.Context()), Message: msg.Content, History: history, UserContext: ctx})
	if e != nil {
		return httpapi.Problem(502, "generation_unavailable", "Не удалось получить ответ. Сообщение сохранено, повторите генерацию")
	}
	if strings.TrimSpace(answer.Answer) == "" || len(answer.Answer) > 32000 || (answer.Mode != "mock" && answer.Mode != "rag") {
		return httpapi.Problem(502, "invalid_rag_response", "Некорректный ответ AI-сервиса")
	}
	// Fail closed: an ML integration may only cite documents authorized by Info.
	for _, source := range answer.Sources {
		if !slices.Contains(ctx.AllowedDocumentIDs, source.DocumentID) {
			return httpapi.Problem(502, "unauthorized_source", "Ответ содержит недоступный источник и не показан")
		}
	}
	if answer.Sources == nil {
		answer.Sources = []contracts.Source{}
	}
	if a := answer.SuggestedAction; a != nil {
		allowed := slices.ContainsFunc(ctx.Tools, func(tool contracts.Tool) bool { return tool.ID == a.System })
		if a.Type != "REQUEST_ACCESS" || !allowed {
			return httpapi.Problem(502, "invalid_action", "AI предложил недоступное действие")
		}
	}
	sources, _ := json.Marshal(answer.Sources)
	action, _ := json.Marshal(answer.SuggestedAction)
	row, e := tx.Query(r.Context(), "INSERT INTO messages(id,chat_id,role,content,reply_to,sources,suggested_action,mode) VALUES($1,$2,'assistant',$3,$4,$5,$6,$7) RETURNING *", idgen.New(), msg.ChatID, answer.Answer, msg.ID, sources, action, answer.Mode)
	if e != nil {
		return e
	}
	result, e := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[Message])
	if e != nil {
		return e
	}
	if e = tx.Commit(r.Context()); e != nil {
		return e
	}
	httpapi.JSON(w, 201, result)
	return nil
}
