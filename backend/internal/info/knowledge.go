package info

import (
	"net/http"
	"strconv"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (i *Service) getDocuments(r *http.Request) ([]contracts.Document, error) {
	return postgres.Query[contracts.Document](r.Context(), i.DB, "SELECT d.* FROM knowledge_documents d JOIN profiles p ON p.user_id=$1 WHERE d.deleted_at IS NULL AND (d.visibility='company' OR d.team_id=p.team_id) ORDER BY d.id", httpapi.User(r).Subject)
}

func (i *Service) documents(w http.ResponseWriter, r *http.Request) error {
	v, e := i.getDocuments(r)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, map[string]any{"items": v})
	return nil
}

func (i *Service) document(w http.ResponseWriter, r *http.Request) error {
	v, e := postgres.One[contracts.Document](r.Context(), i.DB, "SELECT d.* FROM knowledge_documents d JOIN profiles p ON p.user_id=$1 WHERE d.id=$2 AND d.deleted_at IS NULL AND (d.visibility='company' OR d.team_id=p.team_id)", httpapi.User(r).Subject, r.PathValue("id"))
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, v)
	return nil
}

func (i *Service) export(w http.ResponseWriter, r *http.Request) error {
	// Trusted indexer gets ACL metadata and tombstones, never profiles or progress.
	after := r.URL.Query().Get("after")
	limit := 100
	if s := r.URL.Query().Get("limit"); s != "" {
		v, e := strconv.Atoi(s)
		if e != nil || v < 1 || v > 100 {
			return httpapi.Invalid("limit должен быть от 1 до 100")
		}
		limit = v
	}
	docs, e := postgres.Query[contracts.Document](r.Context(), i.DB, "SELECT * FROM knowledge_documents WHERE id>$1 ORDER BY id LIMIT $2", after, limit+1)
	if e != nil {
		return e
	}
	next := ""
	if len(docs) > limit {
		docs = docs[:limit]
		next = docs[len(docs)-1].ID
	}
	httpapi.JSON(w, 200, map[string]any{"documents": docs, "nextCursor": next, "syncMode": "full_snapshot", "aclRequired": true})
	return nil
}
