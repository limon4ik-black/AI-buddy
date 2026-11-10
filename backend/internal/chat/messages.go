package chat

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (c *Service) messages(w http.ResponseWriter, r *http.Request) error {
	if _, e := c.owned(r); e != nil {
		return e
	}
	after := int64(0)
	if v := r.URL.Query().Get("after"); v != "" {
		n, e := strconv.ParseInt(v, 10, 64)
		if e != nil || n < 0 {
			return httpapi.Invalid("Некорректный курсор")
		}
		after = n
	}
	v, e := postgres.Query[Message](r.Context(), c.DB, "SELECT * FROM messages WHERE chat_id=$1 AND sequence>$2 ORDER BY sequence LIMIT 101", r.PathValue("id"), after)
	if e != nil {
		return e
	}
	var next *int64
	if len(v) > 100 {
		v = v[:100]
		n := v[len(v)-1].Sequence
		next = &n
	}
	httpapi.JSON(w, 200, map[string]any{"items": v, "nextCursor": next})
	return nil
}

func (c *Service) send(w http.ResponseWriter, r *http.Request) error {
	if _, e := c.owned(r); e != nil {
		return e
	}
	v, e := httpapi.Decode[struct {
		Content         string `json:"content"`
		ClientMessageID string `json:"clientMessageId"`
	}](r)
	if e != nil {
		return e
	}
	v.Content = strings.TrimSpace(v.Content)
	if v.Content == "" || len(v.Content) > 8000 || len(v.ClientMessageID) < 16 || len(v.ClientMessageID) > 80 {
		return httpapi.Invalid("Нужны сообщение до 8000 байт и clientMessageId длиной 16–80 символов")
	}
	msg, e := postgres.One[Message](r.Context(), c.DB, `INSERT INTO messages(id,chat_id,role,content,client_message_id) VALUES($1,$2,'user',$3,$4)
 ON CONFLICT(chat_id,client_message_id) DO UPDATE SET client_message_id=EXCLUDED.client_message_id RETURNING *`, idgen.New(), r.PathValue("id"), v.Content, v.ClientMessageID)
	if e != nil {
		return e
	}
	if msg.Content != v.Content {
		return httpapi.Problem(409, "idempotency_conflict", "clientMessageId уже использован для другого сообщения")
	}
	if _, e = c.DB.Exec(r.Context(), "UPDATE chats SET updated_at=now() WHERE id=$1", r.PathValue("id")); e != nil {
		return e
	}
	httpapi.JSON(w, 201, msg)
	return nil
}
