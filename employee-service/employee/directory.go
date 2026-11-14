package employee

import (
	"net/http"

	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/postgres"
)

func (i *Service) teams(w http.ResponseWriter, r *http.Request) error {
	v, e := postgres.Query[contracts.Team](r.Context(), i.DB, "SELECT * FROM team_directory ORDER BY id")
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, map[string]any{"items": v, "synthetic": true, "visibility": "company"})
	return nil
}

func (i *Service) getTeam(r *http.Request, id string) (contracts.TeamDetail, error) {
	t, e := postgres.One[contracts.Team](r.Context(), i.DB, "SELECT * FROM team_directory WHERE id=$1", id)
	if e != nil {
		return contracts.TeamDetail{}, e
	}
	members, e := postgres.Query[contracts.Member](r.Context(), i.DB, "SELECT p.*,m.role FROM people p JOIN team_members m ON m.person_id=p.id WHERE m.team_id=$1 ORDER BY m.role,p.id", id)
	if e != nil {
		return contracts.TeamDetail{}, e
	}
	services, e := postgres.Query[contracts.TechnicalService](r.Context(), i.DB, "SELECT * FROM technical_services WHERE owner_team_id=$1 ORDER BY id", id)
	return contracts.TeamDetail{Team: t, Members: members, Services: services}, e
}

func (i *Service) team(w http.ResponseWriter, r *http.Request) error {
	v, e := i.getTeam(r, r.PathValue("id"))
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, v)
	return nil
}

func (i *Service) services(w http.ResponseWriter, r *http.Request) error {
	teamID := r.URL.Query().Get("ownerTeamId")
	if teamID != "" {
		if _, e := postgres.One[contracts.Team](r.Context(), i.DB, "SELECT * FROM team_directory WHERE id=$1", teamID); e != nil {
			return e
		}
	}
	v, e := postgres.Query[contracts.TechnicalService](r.Context(), i.DB, "SELECT * FROM technical_services WHERE ($1='' OR owner_team_id=$1) ORDER BY id", teamID)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, map[string]any{"items": v, "synthetic": true, "visibility": "company"})
	return nil
}

func (i *Service) technicalService(w http.ResponseWriter, r *http.Request) error {
	v, e := postgres.One[contracts.TechnicalService](r.Context(), i.DB, "SELECT * FROM technical_services WHERE id=$1", r.PathValue("id"))
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, v)
	return nil
}
