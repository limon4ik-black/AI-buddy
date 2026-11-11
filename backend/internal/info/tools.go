package info

import (
	"net/http"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (i *Service) getTools(r *http.Request) ([]contracts.Tool, error) {
	return postgres.Query[contracts.Tool](r.Context(), i.DB, "SELECT t.* FROM corporate_tools t JOIN profiles p ON p.user_id=$1 WHERE t.team_id IS NULL OR t.team_id=p.team_id ORDER BY t.id", httpapi.User(r).Subject)
}
func (i *Service) tools(w http.ResponseWriter, r *http.Request) error {
	tools, err := i.getTools(r)
	if err != nil {
		return err
	}
	httpapi.JSON(w, 200, map[string]any{"items": tools})
	return nil
}
