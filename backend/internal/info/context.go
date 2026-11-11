package info

import (
	"net/http"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (i *Service) context(w http.ResponseWriter, r *http.Request) error {
	p, e := i.getProfile(r)
	if e != nil {
		return e
	}
	team, e := i.getTeam(r, p.TeamID)
	if e != nil {
		return e
	}
	list, e := i.getChecklist(r)
	if e != nil {
		return e
	}
	docs, e := i.getDocuments(r)
	if e != nil {
		return e
	}
	ids := []string{}
	for _, d := range docs {
		ids = append(ids, d.ID)
	}
	teams, e := postgres.Query[contracts.Team](r.Context(), i.DB, "SELECT * FROM team_directory ORDER BY id")
	if e != nil {
		return e
	}
	for _, t := range teams {
		ids = append(ids, "team:"+t.ID)
	}
	services, e := postgres.Query[contracts.TechnicalService](r.Context(), i.DB, "SELECT * FROM technical_services ORDER BY id")
	if e != nil {
		return e
	}
	for _, s := range services {
		ids = append(ids, "service:"+s.ID)
	}
	tools, e := i.getTools(r)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, contracts.UserContext{Tools: tools, Profile: p, Team: team, Role: httpapi.User(r).Role, Onboarding: list, AllowedDocumentIDs: ids})
	return nil
}
