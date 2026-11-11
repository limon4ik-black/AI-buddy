package chat

import (
	"net/http"
	"strings"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (c *Service) owned(r *http.Request) (Chat, error) {
	return postgres.One[Chat](r.Context(), c.DB, "SELECT * FROM chats WHERE id=$1 AND user_id=$2", r.PathValue("id"), httpapi.User(r).Subject)
}

func (c *Service) list(w http.ResponseWriter, r *http.Request) error {
	v, e := postgres.Query[Chat](r.Context(), c.DB, "SELECT * FROM chats WHERE user_id=$1 ORDER BY updated_at DESC,id LIMIT 100", httpapi.User(r).Subject)
	if e != nil {
		return e
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
	v, e := postgres.One[Chat](r.Context(), c.DB, "INSERT INTO chats(id,user_id,title) VALUES($1,$2,$3) RETURNING *", idgen.New(), httpapi.User(r).Subject, title)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 201, v)
	return nil
}

func (c *Service) get(w http.ResponseWriter, r *http.Request) error {
	v, e := c.owned(r)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, v)
	return nil
}

func (c *Service) rename(w http.ResponseWriter, r *http.Request) error {
	title, e := chatTitle(r)
	if e != nil {
		return e
	}
	v, e := postgres.One[Chat](r.Context(), c.DB, "UPDATE chats SET title=$3,updated_at=now() WHERE id=$1 AND user_id=$2 RETURNING *", r.PathValue("id"), httpapi.User(r).Subject, title)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, v)
	return nil
}

func (c *Service) delete(w http.ResponseWriter, r *http.Request) error {
	v, e := c.DB.Exec(r.Context(), "DELETE FROM chats WHERE id=$1 AND user_id=$2", r.PathValue("id"), httpapi.User(r).Subject)
	if e != nil {
		return e
	}
	if v.RowsAffected() == 0 {
		return httpapi.NotFound()
	}
	httpapi.JSON(w, 204, nil)
	return nil
}
