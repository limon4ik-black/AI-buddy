package ragmock

import (
	"fmt"
	"slices"
	"strings"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
)

// Answer is a deterministic, permission-aware stand-in for the future RAG service.
func Answer(v contracts.GenerateInput) contracts.GenerateOutput {
	q := strings.ToLower(v.Message)
	ctx := v.UserContext
	out := contracts.GenerateOutput{Mode: "mock", Sources: []contracts.Source{}}
	source := func(id, title string) bool {
		if !slices.Contains(ctx.AllowedDocumentIDs, id) {
			return false
		}
		out.Sources = append(out.Sources, contracts.Source{DocumentID: id, Title: title, Section: "Демонстрационные материалы"})
		return true
	}
	access := strings.Contains(q, "доступ") || strings.Contains(q, "инструмент") || strings.Contains(q, "подключ")
	var chosen *contracts.Tool
	for _, tool := range ctx.Tools {
		if strings.Contains(q, strings.ToLower(tool.ID)) || strings.Contains(q, strings.ToLower(tool.Name)) {
			t := tool
			chosen = &t
			access = true
			break
		}
	}
	if access && chosen == nil {
		for _, name := range []string{"stash", "gitlab", "figma", "storybook", "metabase", "jupyter", "testrail", "тестовый стенд", "vpn", "confluence"} {
			if strings.Contains(q, name) {
				out.Answer = "В доступных демо-материалах нет ответа о доступе к этому инструменту для твоей команды. Уточни его у бадди."
				return out
			}
		}
	}

	if access {
		if chosen == nil {
			for _, tool := range ctx.Tools {
				if tool.TeamID != nil && *tool.TeamID == ctx.Profile.TeamID {
					t := tool
					chosen = &t
					break
				}
			}
		}
		if chosen == nil && len(ctx.Tools) > 0 {
			chosen = &ctx.Tools[0]
		}
		if chosen != nil && source(chosen.DocumentID, "Доступ к "+chosen.Name) {
			out.Answer = fmt.Sprintf("%s, для доступа к %s можно отправить демо-заявку.\n\n1. Нажми «Запросить доступ» и подтверди отправку.\n2. Я сохраню номер заявки в твоём профиле.\n3. Уточни согласование у бадди — %s.\n\nПосле подтверждения заявка появится в профиле со статусом «Отправлена». Сам доступ согласует бадди.", ctx.Profile.FullName, chosen.Name, ctx.Team.Buddy)
			out.SuggestedAction = &contracts.Action{Type: "REQUEST_ACCESS", Label: "Запросить доступ", System: chosen.ID}
		}
	} else if strings.TrimSpace(q) == "задачи" || strings.Contains(q, "остал") || strings.Contains(q, "прогресс") || strings.Contains(q, "следующ") || strings.Contains(q, "чек-лист") {
		out.Answer = fmt.Sprintf("Выполнено %d из %d шагов адаптации — %d%%.", ctx.Onboarding.Completed, ctx.Onboarding.Total, ctx.Onboarding.Percent)
		if ctx.Onboarding.NextTask != nil {
			out.Answer += "\n\nСледующий шаг: " + ctx.Onboarding.NextTask.Title + ".\n" + ctx.Onboarding.NextTask.Description + "\n\nОтметку можно изменить в профиле, когда закончишь."
		} else {
			out.Answer += "\n\nВсе выбранные этапы завершены. Обсуди дальнейшие цели с бадди."
		}
		source("welcome", "Первые недели в команде")
	} else if strings.Contains(q, "сервис") || strings.Contains(q, "отвеча") || strings.Contains(q, "владел") {
		if source("team:"+ctx.Team.ID, "Сервисы команды "+ctx.Team.Name) {
			var lines []string
			for _, svc := range ctx.Team.Services {
				owner := svc.TechnicalOwnerID
				for _, member := range ctx.Team.Members {
					if member.ID == svc.TechnicalOwnerID {
						owner = member.FullName
					}
				}
				lines = append(lines, "• "+svc.Name+" — "+owner+". "+svc.Description)
			}
			out.Answer = "В зоне ответственности команды " + ctx.Team.Name + ":\n\n" + strings.Join(lines, "\n\n") + "\n\nСервисы других команд можно найти в общем каталоге."
		}
	} else if strings.Contains(q, "инструкц") || strings.Contains(q, "принцип") || strings.Contains(q, "первую задач") || strings.Contains(q, "разработ") {
		id, title := "learning-plan", "Личный план обучения"
		switch ctx.Profile.TeamID {
		case "backend":
			id, title = "engineering", "Принципы разработки Backend"
		case "frontend":
			id, title = "frontend-components", "Компоненты и доступность"
		case "analytics":
			id, title = "analytics-metrics", "Словарь метрик"
		case "qa":
			id, title = "qa-test-design", "Тест-дизайн первого сценария"
		}
		if source(id, title) {
			out.Answer = fmt.Sprintf("Для команды %s начни с материала «%s». Открой источник под ответом, затем обсуди вопросы с бадди: %s.\n\nПервые задачи и этапы адаптации собраны в твоём профиле.", ctx.Team.Name, title, ctx.Team.Buddy)
		}
	} else if strings.Contains(q, "команд") || strings.Contains(q, "бадди") || strings.Contains(q, "настав") || strings.Contains(q, "привет") {
		if source("team:"+ctx.Team.ID, "Команда "+ctx.Team.Name) || source("welcome", "Первые недели в команде") {
			var people []string
			for _, member := range ctx.Team.Members {
				if member.Role != "intern" {
					people = append(people, member.FullName+" — "+member.Position)
				}
			}
			out.Answer = fmt.Sprintf("Привет, %s! Ты — %s в команде %s.\n\nТвой бадди — %s. С ним можно обсудить первые задачи, доступы и план обучения.\n\nКоманда рядом:\n%s\n\nНачни со знакомства с коллегами и загляни в свой чек-лист.", ctx.Profile.FullName, ctx.Profile.Position, ctx.Team.Name, ctx.Team.Buddy, strings.Join(people, "\n"))
		}
	}
	if out.Answer == "" {
		out.Answer = "В доступных демо-материалах нет ответа на этот вопрос. Уточни его у бадди.\n\nПока я отвечаю в демо-режиме: могу рассказать о твоей команде, её сервисах, доступах, инструкциях и следующем шаге адаптации."
		out.SuggestedAction = nil
	}
	return out
}
