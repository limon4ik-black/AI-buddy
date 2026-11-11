package chat

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/security"
)

type Service struct {
	DB                                                 *pgxpool.Pool
	Keys                                               *security.Keys
	GatewayToken, ChatToken, RAGToken, InfoURL, RAGURL string
	Client                                             *http.Client
}

func (c *Service) Handler() http.Handler {
	m := http.NewServeMux()
	httpapi.Health(m, c.DB, "chat")
	add := func(pattern string, fn httpapi.Endpoint) {
		m.HandleFunc(pattern, httpapi.Handle(httpapi.ServiceToken(c.GatewayToken, httpapi.Authenticated(c.Keys, fn))))
	}
	add("GET /api/v1/chats", c.list)
	add("POST /api/v1/chats", c.create)
	add("GET /api/v1/chats/{id}", c.get)
	add("PATCH /api/v1/chats/{id}", c.rename)
	add("DELETE /api/v1/chats/{id}", c.delete)
	add("GET /api/v1/chats/{id}/messages", c.messages)
	add("POST /api/v1/chats/{id}/messages", c.send)
	add("POST /api/v1/chats/{id}/messages/{messageId}/answer", c.answer)
	return httpapi.Middleware(m)
}
