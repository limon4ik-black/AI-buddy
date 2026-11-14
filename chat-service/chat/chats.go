package chat

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/postgres"
)

func (c *Service) owned(r *http.Request) (Chat, error) {
	if _, err := uuid.Parse(r.PathValue("id")); err != nil {
		return Chat{}, httpapi.NotFound()
	}
	item, err := postgres.One[Chat](r.Context(), c.DB, "SELECT id::text AS id,user_id::text AS user_id,title,created_at,updated_at FROM chats WHERE id=$1 AND user_id=$2", r.PathValue("id"), owner(httpapi.User(r).Subject))
	item.UserID = httpapi.User(r).Subject
	return item, err
}

func (c *Service) list(w http.ResponseWriter, r *http.Request) error {
	v, e := postgres.Query[Chat](r.Context(), c.DB, "SELECT id::text AS id,user_id::text AS user_id,title,created_at,updated_at FROM chats WHERE user_id=$1 ORDER BY updated_at DESC,id LIMIT 100", owner(httpapi.User(r).Subject))
	if e != nil {
		return e
	}
	for i := range v {
		v[i].UserID = httpapi.User(r).Subject
	}
	httpapi.JSON(w, 200, map[string]any{"items": v, "limit": 100})
	return nil
}

func chatTitle(r *http.Request) (string, error) {
	v, e := httpapi.Decode[struct {
		Title string `json:"title"`
	}](r)
	if e != nil {
		return "", e
	}
	title := strings.TrimSpace(v.Title)
	if len(title) < 1 || len(title) > 200 {
		return "", httpapi.Invalid("Название должно содержать от 1 до 200 байт")
	}
	return title, nil
}

func (c *Service) create(w http.ResponseWriter, r *http.Request) error {
	title, e := chatTitle(r)
	if e != nil {
		return e
	}
	v, e := postgres.One[Chat](r.Context(), c.DB, "INSERT INTO chats(id,user_id,title) VALUES($1,$2,$3) RETURNING id::text AS id,user_id::text AS user_id,title,created_at,updated_at", uuid.New().String(), owner(httpapi.User(r).Subject), title)
	if e != nil {
		return e
	}
	v.UserID = httpapi.User(r).Subject
	httpapi.JSON(w, 201, v)
	return nil
}

func (c *Service) get(w http.ResponseWriter, r *http.Request) error {
	v, e := c.owned(r)
	if e != nil {
		return e
	}
	v.UserID = httpapi.User(r).Subject
	httpapi.JSON(w, 200, v)
	return nil
}

func (c *Service) rename(w http.ResponseWriter, r *http.Request) error {
	if _, err := c.owned(r); err != nil {
		return err
	}
	title, e := chatTitle(r)
	if e != nil {
		return e
	}
	v, e := postgres.One[Chat](r.Context(), c.DB, "UPDATE chats SET title=$3,updated_at=now() WHERE id=$1 AND user_id=$2 RETURNING id::text AS id,user_id::text AS user_id,title,created_at,updated_at", r.PathValue("id"), owner(httpapi.User(r).Subject), title)
	if e != nil {
		return e
	}
	v.UserID = httpapi.User(r).Subject
	httpapi.JSON(w, 200, v)
	return nil
}

func (c *Service) delete(w http.ResponseWriter, r *http.Request) error {
	if _, err := c.owned(r); err != nil {
		return err
	}
	v, e := c.DB.Exec(r.Context(), "DELETE FROM chats WHERE id=$1 AND user_id=$2", r.PathValue("id"), owner(httpapi.User(r).Subject))
	if e != nil {
		return e
	}
	if v.RowsAffected() == 0 {
		return httpapi.NotFound()
	}
	httpapi.JSON(w, 204, nil)
	return nil
}
