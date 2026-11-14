package ragmock

import (
	"net/http"
	"strings"

	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
)

func New(secret string) http.Handler {
	m := http.NewServeMux()
	httpapi.Health(m, nil, "rag-mock")
	m.HandleFunc("POST /internal/v1/generate", httpapi.Handle(httpapi.ServiceToken(secret, func(w http.ResponseWriter, r *http.Request) error {
		v, e := httpapi.Decode[contracts.GenerateInput](r)
		if e != nil {
			return e
		}
		if strings.TrimSpace(v.Message) == "" || len(v.Message) > 8000 || len(v.History) > 20 || v.UserContext.Profile.UserID == "" {
			return httpapi.Invalid("Нужны вопрос и авторизованный контекст")
		}
		if v.UserID == "" || v.UserID != v.UserContext.Profile.UserID {
			return httpapi.Invalid("userId должен совпадать с Auth ID в профиле")
		}
		httpapi.JSON(w, 200, Answer(v))
		return nil
	})))
	return httpapi.Middleware(m)
}
