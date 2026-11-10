package ragmock

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
)

func TestMockBehavior(t *testing.T) {
	ctx := contracts.UserContext{Profile: contracts.Profile{UserID: "u", FullName: "Евгений", Position: "Go Intern", TeamID: "backend"}, Team: contracts.TeamDetail{Team: contracts.Team{Name: "Backend", Buddy: "Алекс"}}, AllowedDocumentIDs: []string{"welcome", "access-backend", "engineering"}, Onboarding: contracts.Checklist{Total: 6, Completed: 2, Percent: 33, NextTask: &contracts.Task{Title: "Next"}}}
	for _, q := range []string{"Привет", "Расскажи о команде", "доступ Stash", "доступ Confluence", "VPN", "Принципы разработки", "следующий шаг", "неизвестный вопрос"} {
		a := Answer(contracts.GenerateInput{Message: q, UserContext: ctx})
		if a.Answer == "" || a.Mode != "mock" {
			t.Fatal(a)
		}
	}
	ctx.Onboarding.NextTask = nil
	a := Answer(contracts.GenerateInput{Message: "задачи", UserContext: ctx})
	if !strings.Contains(a.Answer, "Все выбранные этапы") {
		t.Fatal(a)
	}
	ctx.AllowedDocumentIDs = nil
	for _, q := range []string{"команда", "доступ", "принципы"} {
		a := Answer(contracts.GenerateInput{Message: q, UserContext: ctx})
		if len(a.Sources) > 0 || a.SuggestedAction != nil || !strings.Contains(a.Answer, "нет ответа") {
			t.Fatal("Mock leaked restricted content")
		}
	}
	mock := httptest.NewServer(New("secret"))
	defer mock.Close()
	req(t, mock.Client(), mock.URL, "POST", "/internal/v1/generate", "", contracts.GenerateInput{}, 400, "X-Service-Token", "secret")
	req(t, mock.Client(), mock.URL, "POST", "/internal/v1/generate", "", contracts.GenerateInput{}, 401)
	req(t, mock.Client(), mock.URL, "POST", "/internal/v1/generate", "", map[string]any{"invalid": true}, 400, "X-Service-Token", "secret")
	req(t, mock.Client(), mock.URL, "GET", "/healthz", "", nil, 200)
}
