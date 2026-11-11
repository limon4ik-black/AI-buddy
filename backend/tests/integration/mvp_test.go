package integration

import (
	"context"
	"slices"
	"testing"

	chatsvc "github.com/limon4ik-black/AI-buddy/backend/internal/chat"
	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"github.com/limon4ik-black/AI-buddy/backend/internal/demodata"
	"github.com/limon4ik-black/AI-buddy/backend/internal/info"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
)

func TestMVPDemoTeams(t *testing.T) {
	h := setup(t)
	h.auth.Mailer = &captureMailer{}
	if err := h.info.SeedDemo(context.Background()); err != nil {
		t.Fatal(err)
	}
	tokens := map[string]string{}
	chats := map[string]string{}
	for _, u := range demodata.Interns() {
		if u.ID == "demo-intern" {
			continue
		}
		t.Run(u.TeamID+"/"+u.ID, func(t *testing.T) {
			token := loginSession(t, h, u.Email, demodata.Password)
			tokens[u.ID] = token
			profile := unpack[contracts.Profile](t, req(t, h.client, h.gateway, "GET", "/api/v1/profile", token, nil, 200))
			if profile.UserID != u.ID || profile.TeamID != u.TeamID || profile.FullName != u.FullName {
				t.Fatal(profile)
			}
			list := unpack[contracts.Checklist](t, req(t, h.client, h.gateway, "GET", "/api/v1/checklist", token, nil, 200))
			if list.Total != 15 || list.Completed != len(u.Completed) {
				t.Fatal("incorrect personal progress", list.Total, list.Completed)
			}
			docs := unpack[struct{ Items []contracts.Document }](t, req(t, h.client, h.gateway, "GET", "/api/v1/information", token, nil, 200))
			for _, d := range docs.Items {
				if d.Visibility == "team" && (d.TeamID == nil || *d.TeamID != u.TeamID) {
					t.Fatal("document leaked", d.ID)
				}
			}
			tools := unpack[struct{ Items []contracts.Tool }](t, req(t, h.client, h.gateway, "GET", "/api/v1/tools", token, nil, 200))
			var own contracts.Tool
			for _, tool := range tools.Items {
				if tool.TeamID != nil {
					if *tool.TeamID != u.TeamID {
						t.Fatal("tool leaked")
					}
					own = tool
					break
				}
			}
			if own.ID == "" {
				t.Fatal("no team-specific tool")
			}
			chat := unpack[chatsvc.Chat](t, req(t, h.client, h.gateway, "POST", "/api/v1/chats", token, obj("title", "Доступы "+u.TeamID), 201))
			chats[u.ID] = chat.ID
			path := "/api/v1/chats/" + chat.ID
			msg := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", path+"/messages", token, obj("content", "Как получить доступ к "+own.Name+"?", "clientMessageId", idgen.New()), 201))
			answer := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", path+"/messages/"+msg.ID+"/answer", token, nil, 201))
			if answer.SuggestedAction == nil || answer.SuggestedAction.System != own.ID {
				t.Fatal(answer)
			}
			result := unpack[info.AccessRequest](t, req(t, h.client, h.gateway, "POST", "/api/v1/access-requests", token, obj("system", own.ID, "confirmed", true), 201, "Idempotency-Key", idgen.New()))
			if result.UserID != u.ID || result.Status != "submitted" {
				t.Fatal(result)
			}
			ctx := unpack[contracts.UserContext](t, req(t, h.client, h.infoURL, "GET", "/internal/v1/context", token, nil, 200, "X-Service-Token", "chat-secret"))
			if !slices.ContainsFunc(ctx.Tools, func(tool contracts.Tool) bool { return tool.ID == own.ID }) {
				t.Fatal("context omitted allowed tools")
			}
		})
	}
	for user, token := range tokens {
		requests := unpack[struct{ Items []info.AccessRequest }](t, req(t, h.client, h.gateway, "GET", "/api/v1/access-requests", token, nil, 200))
		if len(requests.Items) != 1 || requests.Items[0].UserID != user {
			t.Fatal("requests leaked")
		}
		for owner, chatID := range chats {
			if owner != user {
				req(t, h.client, h.gateway, "GET", "/api/v1/chats/"+chatID, token, nil, 404)
			}
		}
	}
	egor := tokens["intern-egor"]
	req(t, h.client, h.gateway, "PATCH", "/api/v1/checklist/tasks/meet-team", egor, obj("status", "pending"), 200)
	req(t, h.client, h.gateway, "PATCH", "/api/v1/profile", egor, obj("fullName", "Егор Тестовый", "position", "Go Intern"), 200)
	if err := h.info.SeedDemo(context.Background()); err != nil {
		t.Fatal(err)
	}
	list := unpack[contracts.Checklist](t, req(t, h.client, h.gateway, "GET", "/api/v1/checklist", egor, nil, 200))
	if list.Completed != 2 {
		t.Fatal("reseed overwrote progress")
	}
	profile := unpack[contracts.Profile](t, req(t, h.client, h.gateway, "GET", "/api/v1/profile", egor, nil, 200))
	if profile.FullName != "Егор Тестовый" {
		t.Fatal("reseed overwrote profile")
	}
	req(t, h.client, h.gateway, "GET", "/api/v1/information/access-backend", tokens["intern-anna"], nil, 404)
	req(t, h.client, h.gateway, "POST", "/api/v1/access-requests", tokens["intern-anna"], obj("system", "stash", "confirmed", true), 403, "Idempotency-Key", idgen.New())
}
