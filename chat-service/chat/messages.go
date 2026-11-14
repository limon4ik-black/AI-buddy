package chat

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/postgres"
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
	v, e := postgres.Query[Message](r.Context(), c.DB, "SELECT id::text AS id,chat_id::text AS chat_id,role,content,client_message_id,reply_to::text AS reply_to,sources,suggested_action,mode,id AS sequence,created_at FROM messages WHERE chat_id=$1 AND id>$2 ORDER BY id LIMIT 101", r.PathValue("id"), after)
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
	msg, err := c.saveMessage(r)
	if err != nil {
		return err
	}
	httpapi.JSON(w, 201, msg)
	return nil
}

func (c *Service) saveMessage(r *http.Request) (Message, error) {
	if _, e := c.owned(r); e != nil {
		return Message{}, e
	}
	v, e := httpapi.Decode[struct {
		Content         string `json:"content"`
		ClientMessageID string `json:"clientMessageId"`
	}](r)
	if e != nil {
		return Message{}, e
	}
	v.Content = strings.TrimSpace(v.Content)
	if v.Content == "" || len(v.Content) > 8000 || len(v.ClientMessageID) < 16 || len(v.ClientMessageID) > 80 {
		return Message{}, httpapi.Invalid("Нужны сообщение до 8000 байт и clientMessageId длиной 16–80 символов")
	}
	msg, e := postgres.One[Message](r.Context(), c.DB, `INSERT INTO messages(chat_id,role,content,client_message_id) VALUES($1,'user',$2,$3)
 ON CONFLICT(chat_id,client_message_id) DO UPDATE SET client_message_id=EXCLUDED.client_message_id RETURNING id::text AS id,chat_id::text AS chat_id,role,content,client_message_id,reply_to::text AS reply_to,sources,suggested_action,mode,id AS sequence,created_at`, r.PathValue("id"), v.Content, v.ClientMessageID)
	if e != nil {
		return Message{}, e
	}
	if msg.Content != v.Content {
		return Message{}, httpapi.Problem(409, "idempotency_conflict", "clientMessageId уже использован для другого сообщения")
	}
	if _, e = c.DB.Exec(r.Context(), "UPDATE chats SET updated_at=now() WHERE id=$1", r.PathValue("id")); e != nil {
		return Message{}, e
	}
	return msg, nil
}
