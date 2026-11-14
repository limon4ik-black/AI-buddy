package corpmock

import (
	"net/http"
	"strings"

	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/demodata"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/shared/platform/security"
)

func New(secret string) http.Handler {
	m := http.NewServeMux()
	httpapi.Health(m, nil, "corp-mock")
	m.HandleFunc("POST /internal/v1/access-requests", httpapi.Handle(httpapi.ServiceToken(secret, func(w http.ResponseWriter, r *http.Request) error {
		v, e := httpapi.Decode[contracts.CorpInput](r)
		if e != nil {
			return e
		}
		if v.UserID == "" || len(v.IdempotencyKey) < 16 || !demodata.SupportsTool(v.System) {
			return httpapi.Invalid("Неизвестный инструмент или некорректный запрос")
		}
		// Deterministic mock provider result survives restarts and duplicate delivery.
		httpapi.JSON(w, 201, contracts.CorpResult{ID: "DEMO-" + strings.ToUpper(security.TokenHash(v.UserID + ":" + v.IdempotencyKey + ":" + v.System)[:12]), Status: "submitted"})
		return nil
	})))
	return httpapi.Middleware(m)
}
