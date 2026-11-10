package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
	"github.com/limon4ik-black/AI-buddy/backend/migrations"
)

func TestInfoAndDirectory(t *testing.T) {
	h := setup(t)
	get := func(path string, status int) []byte {
		return req(t, h.client, h.gateway, "GET", "/api/v1"+path, h.token, nil, status)
	}
	get("/profile", 200)
	p := unpack[contracts.Profile](t, get("/profile", 200))
	if p.TeamID != "backend" || p.UserID != "demo-intern" {
		t.Fatal(p)
	}
	req(t, h.client, h.gateway, "PATCH", "/api/v1/profile", h.token, obj("fullName", "Евгений", "position", "Go Intern"), 200)
	req(t, h.client, h.gateway, "PATCH", "/api/v1/profile", h.token, obj("fullName", "", "position", "Go"), 400)
	req(t, h.client, h.gateway, "PATCH", "/api/v1/profile", h.token, obj("teamId", "analytics"), 400)
	req(t, h.client, h.gateway, "PATCH", "/api/v1/profile", h.token, nil, 400)
	teams := unpack[struct{ Items []contracts.Team }](t, get("/teams", 200))
	if len(teams.Items) != 4 {
		t.Fatal(teams)
	}
	team := unpack[contracts.TeamDetail](t, get("/teams/backend", 200))
	if len(team.Members) != 3 || len(team.Services) != 4 {
		t.Fatal(team)
	}
	get("/teams/analytics", 200)
	get("/teams/nope", 404)
	get("/services", 200)
	get("/services?ownerTeamId=backend", 200)
	get("/services?ownerTeamId=nope", 404)
	get("/services/knowledge-api", 200)
	get("/services/nope", 404)
	get("/tools", 200)
	docs := string(get("/information", 200))
	if strings.Contains(docs, "analytics-private") {
		t.Fatal("Team ACL leaked")
	}
	get("/information/welcome", 200)
	get("/information/analytics-private", 404)
	list := unpack[contracts.Checklist](t, get("/checklist", 200))
	if list.Total != 15 || list.NextTask == nil {
		t.Fatal(list)
	}
	for _, task := range list.Tasks {
		req(t, h.client, h.gateway, "PATCH", "/api/v1/checklist/tasks/"+task.ID, h.token, obj("status", "completed"), 200)
	}
	list = unpack[contracts.Checklist](t, get("/checklist", 200))
	if list.Percent != 100 || list.NextTask != nil {
		t.Fatal(list)
	}
	for _, status := range []string{"completed", "in_progress", "pending"} {
		req(t, h.client, h.gateway, "PATCH", "/api/v1/checklist/tasks/meet-team", h.token, obj("status", status), 200)
	}
	req(t, h.client, h.gateway, "PATCH", "/api/v1/checklist/tasks/nope", h.token, obj("status", "completed"), 404)
	req(t, h.client, h.gateway, "PATCH", "/api/v1/checklist/tasks/meet-team", h.token, obj("status", "nope"), 400)
	req(t, h.client, h.gateway, "PATCH", "/api/v1/checklist/tasks/meet-team", h.token, obj("unknown", "foo"), 400)
	for _, path := range []string{"/teams", "/teams/backend", "/services", "/services?ownerTeamId=analytics", "/services/intern-profile", "/checklist/templates?teamId=backend", "/checklist/templates?teamId=analytics"} {
		req(t, h.client, h.infoURL, "GET", "/internal/v1"+path, "", nil, 200, "X-Service-Token", "index-secret")
	}
	for _, path := range []string{"/checklist/templates", "/checklist/templates?teamId=nope"} {
		want := 400
		if strings.Contains(path, "nope") {
			want = 404
		}
		req(t, h.client, h.infoURL, "GET", "/internal/v1"+path, "", nil, want, "X-Service-Token", "index-secret")
	}
	req(t, h.client, h.infoURL, "GET", "/internal/v1/teams", "", nil, 401)
	req(t, h.client, h.infoURL, "GET", "/internal/v1/context", h.token, nil, 401, "X-Service-Token", "index-secret")
	ctx := unpack[contracts.UserContext](t, req(t, h.client, h.infoURL, "GET", "/internal/v1/context", h.token, nil, 200, "X-Service-Token", "chat-secret"))
	if ctx.Profile.FullName != "Евгений" || ctx.Onboarding.Completed != 14 || len(ctx.Team.Members) != 3 {
		t.Fatal(ctx)
	}
	cursor := ""
	count := 0
	for {
		page := unpack[struct {
			Documents  []contracts.Document
			NextCursor string
		}](t, req(t, h.client, h.infoURL, "GET", "/internal/v1/knowledge/documents?limit=1&after="+cursor, "", nil, 200, "X-Service-Token", "index-secret"))
		count += len(page.Documents)
		cursor = page.NextCursor
		if cursor == "" {
			break
		}
	}
	if count != 30 {
		t.Fatal(count)
	}
	req(t, h.client, h.infoURL, "GET", "/internal/v1/knowledge/documents", "", nil, 200, "X-Service-Token", "index-secret")
	for _, limit := range []string{"0", "101", "bad"} {
		req(t, h.client, h.infoURL, "GET", "/internal/v1/knowledge/documents?limit="+limit, "", nil, 400, "X-Service-Token", "index-secret")
	}
	// Second trainee has isolated progress and a different team assigned by the organization.
	_, e := h.info.DB.Exec(context.Background(), "INSERT INTO profiles(user_id,team_id) VALUES('other','analytics')")
	if e != nil {
		t.Fatal(e)
	}
	other := h.keys.Sign("other", "other@example.test")
	otherList := unpack[contracts.Checklist](t, req(t, h.client, h.gateway, "GET", "/api/v1/checklist", other, nil, 200))
	if otherList.Completed != 0 {
		t.Fatal("Progress leaked")
	}
	req(t, h.client, h.gateway, "PATCH", "/api/v1/checklist/tasks/stash-keys", other, obj("status", "completed"), 404)
	req(t, h.client, h.gateway, "GET", "/api/v1/information/analytics-private", other, nil, 200)
	req(t, h.client, h.gateway, "GET", "/api/v1/information/access-backend", other, nil, 404)
	_, e = h.info.DB.Exec(context.Background(), "INSERT INTO technical_services(id,name,description,owner_team_id,technical_owner_id,lifecycle,repository_url,documentation_url) VALUES('bad','bad','bad','backend','sasha','active','','')")
	if e == nil {
		t.Fatal("Accepted owner from another team")
	}
	if e = postgres.Migrate(context.Background(), h.info.DB, "info", migrations.Info); e != nil {
		t.Fatal("Migration is not idempotent", e)
	}
}
