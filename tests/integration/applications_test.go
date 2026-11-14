package integration

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/limon4ik-black/AI-buddy/auth-service/auth"
	"github.com/limon4ik-black/AI-buddy/chat-service/chat"
	"github.com/limon4ik-black/AI-buddy/corpmock-service/corpmock"
	"github.com/limon4ik-black/AI-buddy/gateway-service/gateway"
	info "github.com/limon4ik-black/AI-buddy/employee-service/employee"
	"github.com/limon4ik-black/AI-buddy/ragmock-service/ragmock"
	"github.com/limon4ik-black/AI-buddy/shared/platform/server"
)

// Exercise production wiring as well as the in-process service handlers.
func TestApplicationFactories(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	t.Setenv("JWT_PRIVATE_SEED", base64.StdEncoding.EncodeToString(priv.Seed()))
	t.Setenv("JWT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(pub))
	for _, key := range []string{"GATEWAY_SERVICE_TOKEN", "CHAT_SERVICE_TOKEN", "INDEXER_SERVICE_TOKEN", "RAG_SERVICE_TOKEN", "CORP_SERVICE_TOKEN"} {
		t.Setenv(key, strings.Repeat("s", 32))
	}
	t.Setenv("SEED_DEMO", "true")
	factories := map[string]server.Factory{"auth": auth.NewApplication, "chat": chat.NewApplication, "info": info.NewApplication, "gateway": gateway.NewApplication, "ragmock": ragmock.NewApplication, "corpmock": corpmock.NewApplication}
	for _, name := range []string{"auth", "chat", "info", "gateway", "ragmock", "corpmock"} {
		t.Run(name, func(t *testing.T) {
			if name == "auth" || name == "chat" || name == "info" {
				db := database(t, name)
				t.Setenv("DATABASE_URL", db.Config().ConnString())
			}
			handler, cleanup, err := factories[name](context.Background())
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()
			srv := httptest.NewServer(handler)
			defer srv.Close()
			req(t, srv.Client(), srv.URL, "GET", "/healthz", "", nil, 200)
		})
	}
}
