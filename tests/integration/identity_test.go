package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	chatsvc "github.com/limon4ik-black/AI-buddy/chat-service/chat"
	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/platform/idgen"
)

func TestIdentityReconcilesExistingAuthIDWithSeededInfo(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	authID := "existing-auth-" + idgen.New()
	if _, err := h.auth.DB.Exec(ctx, "UPDATE users SET id=$1 WHERE email='backend@buddy.test'", authID); err != nil {
		t.Fatal(err)
	}
	if err := h.auth.SeedDemo(ctx); err != nil {
		t.Fatal(err)
	}
	var actualID string
	if err := h.auth.DB.QueryRow(ctx, "SELECT id FROM users WHERE email='backend@buddy.test'").Scan(&actualID); err != nil || actualID != authID {
		t.Fatalf("Auth reseed changed account identity: %q, %v", actualID, err)
	}
	if err := h.info.SeedDemo(ctx); err != nil {
		t.Fatal(err)
	}
	// Repair the empty profile left by the previous implementation too.
	if _, err := h.info.DB.Exec(ctx, "INSERT INTO profiles(user_id) VALUES($1)", authID); err != nil {
		t.Fatal(err)
	}
	token := h.keys.Sign(authID, "backend@buddy.test")
	path := "/internal/v1/users/" + authID + "/context"
	result := unpack[contracts.UserContext](t, req(t, h.client, h.infoURL, "GET", path, token, nil, 200, "X-Service-Token", "chat-secret"))
	if result.Profile.UserID != authID || result.Profile.FullName != "Егор Соколов" || result.Profile.TeamID != "backend" {
		t.Fatalf("canonical profile lost seeded identity: %+v", result.Profile)
	}
	if result.Onboarding.Completed != 3 || result.Onboarding.Total != 15 {
		t.Fatalf("seeded progress lost: %+v", result.Onboarding)
	}
	// The normal JWT-only context and frontend profile must resolve the same person.
	legacy := unpack[contracts.UserContext](t, req(t, h.client, h.infoURL, "GET", "/internal/v1/context", token, nil, 200, "X-Service-Token", "chat-secret"))
	profile := unpack[contracts.Profile](t, req(t, h.client, h.gateway, "GET", "/api/v1/profile", token, nil, 200))
	if legacy.Profile != result.Profile || profile != result.Profile {
		t.Fatal("context APIs resolve different profiles")
	}
	var count int
	if err := h.info.DB.QueryRow(ctx, "SELECT count(*) FROM profiles WHERE user_id=$1", authID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("canonical profile is not unique: count=%d err=%v", count, err)
	}
	foreign := h.keys.Sign("other-user", "other@example.test")
	req(t, h.client, h.infoURL, "GET", path, foreign, nil, 403, "X-Service-Token", "chat-secret")
	req(t, h.client, h.infoURL, "GET", path, token, nil, 401, "X-Service-Token", "index-secret")
	req(t, h.client, h.infoURL, "GET", path, "", nil, 401, "X-Service-Token", "chat-secret")
	// User edits must survive subsequent context reads and seed runs.
	req(t, h.client, h.gateway, "PATCH", "/api/v1/profile", token, obj("fullName", "Егор Изменённый", "position", "Intern"), 200)
	req(t, h.client, h.gateway, "PATCH", "/api/v1/checklist/tasks/meet-team", token, obj("status", "pending"), 200)
	if err := h.info.SeedDemo(ctx); err != nil {
		t.Fatal(err)
	}
	updated := unpack[contracts.UserContext](t, req(t, h.client, h.infoURL, "GET", path, token, nil, 200, "X-Service-Token", "chat-secret"))
	if updated.Profile.FullName != "Егор Изменённый" || updated.Onboarding.Completed != 2 {
		t.Fatal("reconciliation overwrote canonical edits", updated.Profile, updated.Onboarding.Completed)
	}
}

func TestChatForwardsCanonicalAuthIDToRAG(t *testing.T) {
	h := setup(t)
	subject := "registered-" + idgen.New()
	email := subject + "@example.test"
	if _, err := h.auth.DB.Exec(context.Background(), "INSERT INTO users(id,email) VALUES($1,$2)", subject, email); err != nil {
		t.Fatal(err)
	}
	token := h.keys.Sign(subject, email)
	profile := unpack[contracts.Profile](t, req(t, h.client, h.gateway, "GET", "/api/v1/profile", token, nil, 200))
	if profile.UserID != subject {
		t.Fatal("new profile uses a different identity", profile.UserID)
	}
	// Capture the actual JSON contract without relying on native chat owner UUIDs.
	type capturedInput struct {
		UserID      string                `json:"userId"`
		UserContext contracts.UserContext `json:"userContext"`
	}
	captured := make(chan capturedInput, 1)
	rag := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Error("Chat did not forward the original user JWT to RAG")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/internal/v1/generate" || r.Header.Get("X-Service-Token") != "rag-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var input capturedInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		captured <- input
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contracts.GenerateOutput{Answer: "Добро пожаловать", Mode: "rag", Sources: []contracts.Source{}})
	}))
	defer rag.Close()
	h.chat.RAGURL = rag.URL
	chat := unpack[chatsvc.Chat](t, req(t, h.client, h.gateway, "POST", "/api/v1/chats", token, obj("title", "Identity check"), 201))
	if chat.UserID != subject {
		t.Fatal("frontend received internal owner UUID", chat.UserID)
	}
	path := "/api/v1/chats/" + chat.ID
	message := unpack[chatsvc.Message](t, req(t, h.client, h.gateway, "POST", path+"/messages", token, obj("content", "Как начать?", "clientMessageId", idgen.New()), 201))
	req(t, h.client, h.gateway, "POST", path+"/messages/"+message.ID+"/answer", token, nil, 201)
	select {
	case input := <-captured:
		if input.UserID != subject || input.UserContext.Profile.UserID != subject {
			t.Fatalf("RAG received a different identity: top-level=%q profile=%q want=%q", input.UserID, input.UserContext.Profile.UserID, subject)
		}
	default:
		t.Fatal("RAG did not receive generation request")
	}
}

func TestIdentityRepairRemovesStaleDefaultTeamMembership(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	authID := "existing-analytics-" + idgen.New()
	if _, err := h.auth.DB.Exec(ctx, "UPDATE users SET id=$1 WHERE email='analytics@buddy.test'", authID); err != nil {
		t.Fatal(err)
	}
	if err := h.info.SeedDemo(ctx); err != nil {
		t.Fatal(err)
	}
	// Reproduce a previous request that provisioned the generic backend profile.
	tx, err := h.info.DB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	for _, sql := range []string{
		"INSERT INTO profiles(user_id) VALUES($1)",
		"INSERT INTO people(id,full_name,position) VALUES($1,'Стажёр','Backend-стажёр')",
		"INSERT INTO team_members(team_id,person_id,role) VALUES('backend',$1,'intern')",
	} {
		if _, err = tx.Exec(ctx, sql, authID); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	token := h.keys.Sign(authID, "analytics@buddy.test")
	result := unpack[contracts.UserContext](t, req(t, h.client, h.infoURL, "GET", "/internal/v1/users/"+authID+"/context", token, nil, 200, "X-Service-Token", "chat-secret"))
	if result.Profile.UserID != authID || result.Profile.TeamID != "analytics" || result.Profile.FullName != "Максим Орлов" {
		t.Fatalf("analytics identity repair failed: %+v", result.Profile)
	}
	rows, err := h.info.DB.Query(ctx, "SELECT team_id FROM team_members WHERE person_id=$1 AND role='intern' ORDER BY team_id", authID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	teams := []string{}
	for rows.Next() {
		var team string
		if err := rows.Scan(&team); err != nil {
			t.Fatal(err)
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(teams) != 1 || teams[0] != "analytics" {
		t.Fatal("stale default team membership survived repair", teams)
	}
}
