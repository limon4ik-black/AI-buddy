package ragmock

import (
	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"strings"
	"testing"
)

func TestTeamAwareAnswers(t *testing.T) {
	for team, document := range map[string]string{"backend": "engineering", "frontend": "frontend-components", "analytics": "analytics-metrics", "qa": "qa-test-design"} {
		ctx := contracts.UserContext{Profile: contracts.Profile{TeamID: team, FullName: "Стажёр", UserID: "user"}, Team: contracts.TeamDetail{Team: contracts.Team{ID: team, Name: team, Buddy: "Бадди"}, Members: []contracts.Member{{ID: "owner", FullName: "Ответственный", Position: "Инженер"}}, Services: []contracts.TechnicalService{{ID: "svc", Name: "Сервис", TechnicalOwnerID: "owner"}}}, AllowedDocumentIDs: []string{document, "team:" + team, "frontend-access", "welcome"}}
		for _, q := range []string{"Какие инструкции изучить?", "Расскажи о команде", "Кто отвечает за сервисы?"} {
			result := Answer(contracts.GenerateInput{Message: q, UserContext: ctx})
			if len(result.Sources) == 0 || strings.Contains(result.Answer, "нет ответа") {
				t.Fatal(team, q, result)
			}
		}
	}
	team := "frontend"
	ctx := contracts.UserContext{Profile: contracts.Profile{TeamID: team}, Tools: []contracts.Tool{{ID: "figma", Name: "Figma", TeamID: &team, DocumentID: "frontend-access"}}, AllowedDocumentIDs: []string{"frontend-access"}}
	for _, q := range []string{"Доступ к Figma", "Нужные инструменты"} {
		result := Answer(contracts.GenerateInput{Message: q, UserContext: ctx})
		if result.SuggestedAction == nil || result.SuggestedAction.System != "figma" {
			t.Fatal(result)
		}
		if strings.Contains(result.Answer, "зарегистрирована") {
			t.Fatal("claimed an action was completed before confirmation")
		}
	}
	denied := Answer(contracts.GenerateInput{Message: "Как получить доступ к Stash?", UserContext: ctx})
	if denied.SuggestedAction != nil {
		t.Fatal("suggested an unavailable tool")
	}
	ctx.Tools[0].TeamID = nil
	result := Answer(contracts.GenerateInput{Message: "Как получить доступы?", UserContext: ctx})
	if result.SuggestedAction == nil {
		t.Fatal("common tool not offered")
	}
}
