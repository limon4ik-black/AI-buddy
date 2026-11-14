package chat

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpclient"
)

func (c *Service) generate(r *http.Request) (Message, bool, error) {
	if n, err := strconv.ParseInt(r.PathValue("messageId"), 10, 64); err != nil || n < 1 {
		return Message{}, false, httpapi.NotFound()
	}
	if _, e := c.owned(r); e != nil {
		return Message{}, false, e
	}
	tx, e := c.DB.Begin(r.Context())
	if e != nil {
		return Message{}, false, e
	}
	defer tx.Rollback(r.Context())
	if _, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", r.PathValue("id")); e != nil {
		return Message{}, false, e
	}
	var msg Message
	rows, e := tx.Query(r.Context(), "SELECT id::text AS id,chat_id::text AS chat_id,role,content,client_message_id,reply_to::text AS reply_to,sources,suggested_action,mode,id AS sequence,created_at FROM messages WHERE id=$1 AND chat_id=$2 AND role='user'", r.PathValue("messageId"), r.PathValue("id"))
	if e != nil {
		return Message{}, false, e
	}
	msg, e = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Message])
	if e != nil {
		return Message{}, false, e
	}
	existing, e := tx.Query(r.Context(), "SELECT id::text AS id,chat_id::text AS chat_id,role,content,client_message_id,reply_to::text AS reply_to,sources,suggested_action,mode,id AS sequence,created_at FROM messages WHERE reply_to=$1", msg.ID)
	if e != nil {
		return Message{}, false, e
	}
	previous, e := pgx.CollectExactlyOneRow(existing, pgx.RowToStructByName[Message])
	if e == nil {
		return previous, true, nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return Message{}, false, e
	}
	userID := httpapi.User(r).Subject
	ctx, e := httpclient.Call[contracts.UserContext](r.Context(), c.Client, "GET", c.NoteURL+"/internal/v1/users/"+url.PathEscape(userID)+"/context", c.ChatToken, r.Header.Get("Authorization"), nil)
	if e != nil {
		return Message{}, false, httpapi.Problem(502, "context_unavailable", "Не удалось получить актуальный контекст. Сообщение сохранено, повторите генерацию")
	}
	if ctx.Profile.UserID != userID {
		return Message{}, false, httpapi.Problem(502, "identity_mismatch", "Note Service вернул контекст другого пользователя")
	}
	historyRows, e := tx.Query(r.Context(), "SELECT role,content FROM (SELECT role,content,id FROM messages WHERE chat_id=$1 AND id<$2 ORDER BY id DESC LIMIT 20) h ORDER BY id", msg.ChatID, msg.Sequence)
	if e != nil {
		return Message{}, false, e
	}
	history, e := pgx.CollectRows(historyRows, pgx.RowToStructByName[contracts.Turn])
	if e != nil {
		return Message{}, false, e
	}
	answer, e := httpclient.Call[contracts.GenerateOutput](r.Context(), c.Client, "POST", c.RAGURL+"/internal/v1/generate", c.RAGToken, r.Header.Get("Authorization"), contracts.GenerateInput{RequestID: httpapi.RequestID(r.Context()), UserID: userID, Message: msg.Content, History: history, UserContext: ctx})
	if e != nil {
		return Message{}, false, httpapi.Problem(502, "generation_unavailable", "Не удалось получить ответ. Сообщение сохранено, повторите генерацию")
	}
	if strings.TrimSpace(answer.Answer) == "" || len(answer.Answer) > 32000 || (answer.Mode != "mock" && answer.Mode != "rag") {
		return Message{}, false, httpapi.Problem(502, "invalid_rag_response", "Некорректный ответ AI-сервиса")
	}
	// Fail closed: an ML integration may only cite documents authorized by Info.
	for _, source := range answer.Sources {
		if !slices.Contains(ctx.AllowedDocumentIDs, source.DocumentID) {
			return Message{}, false, httpapi.Problem(502, "unauthorized_source", "Ответ содержит недоступный источник и не показан")
		}
	}
	if answer.Sources == nil {
		answer.Sources = []contracts.Source{}
	}
	if a := answer.SuggestedAction; a != nil {
		allowed := slices.ContainsFunc(ctx.Tools, func(tool contracts.Tool) bool { return tool.ID == a.System })
		if a.Type != "REQUEST_ACCESS" || !allowed {
			return Message{}, false, httpapi.Problem(502, "invalid_action", "AI предложил недоступное действие")
		}
	}
	sources, _ := json.Marshal(answer.Sources)
	action, _ := json.Marshal(answer.SuggestedAction)
	row, e := tx.Query(r.Context(), "INSERT INTO messages(chat_id,role,content,reply_to,sources,suggested_action,mode) VALUES($1,'assistant',$2,$3,$4,$5,$6) RETURNING id::text AS id,chat_id::text AS chat_id,role,content,client_message_id,reply_to::text AS reply_to,sources,suggested_action,mode,id AS sequence,created_at", msg.ChatID, answer.Answer, msg.ID, sources, action, answer.Mode)
	if e != nil {
		return Message{}, false, e
	}
	result, e := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[Message])
	if e != nil {
		return Message{}, false, e
	}
	if e = tx.Commit(r.Context()); e != nil {
		return Message{}, false, e
	}
	return result, false, nil
}

func (c *Service) answer(w http.ResponseWriter, r *http.Request) error {
	msg, reused, err := c.generate(r)
	if err != nil {
		return err
	}
	status := 201
	if reused {
		status = 200
	}
	httpapi.JSON(w, status, msg)
	return nil
}
