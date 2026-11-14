package employee

import (
	"net/http"

	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
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
	tools, e := i.getTools(r)
	if e != nil {
		return e
	}
	// Checklist, document ACLs and personal progress are owned by Note Service.
	// Note composes this employee-only context with its onboarding data.
	httpapi.JSON(w, 200, contracts.UserContext{Tools: tools, Profile: p, Team: team, Role: httpapi.User(r).Role})
	return nil
}
