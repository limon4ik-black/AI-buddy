package chat

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	nativechat "github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/store"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/transport"
	"github.com/limon4ik-black/AI-buddy/shared/platform/config"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
	"net/http"
	"net/url"
	"os"
	"time"
)

func NewApplication(ctx context.Context) (http.Handler, func(), error) {
	keys, err := config.JWT(false)
	if err != nil {
		return nil, nil, err
	}
	tokens, err := config.Secrets("GATEWAY_SERVICE_TOKEN", "CHAT_SERVICE_TOKEN", "RAG_SERVICE_TOKEN")
	if err != nil {
		return nil, nil, err
	}
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, nil, err
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = "chat_v2"
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	if err = Migrate(ctx, db); err != nil {
		db.Close()
		return nil, nil, err
	}
	service := &Service{DB: db, Keys: keys, GatewayToken: tokens["GATEWAY_SERVICE_TOKEN"], ChatToken: tokens["CHAT_SERVICE_TOKEN"], RAGToken: tokens["RAG_SERVICE_TOKEN"], NoteURL: config.Env("NOTE_URL", "http://note_service:8079"), RAGURL: config.Env("RAG_URL", "http://ragmock:8084"), Client: &http.Client{Timeout: 180 * time.Second}}
	native := &transport.Handler{Service: &nativechat.Service{Repository: &store.Postgres{Pool: db}}, Timeout: 180 * time.Second, SendAuthenticated: service.SendWebSocket}
	if origin, e := url.Parse(config.Env("APP_ORIGIN", "http://localhost:8080")); e == nil && origin.Host != "" {
		native.Origins = []string{origin.Host}
	}
	mux := http.NewServeMux()
	mux.Handle("/api/v1/", service.Handler())
	mux.Handle("GET /healthz", service.Handler())
	mux.Handle("/", httpapi.Handle(httpapi.ServiceToken(service.GatewayToken, httpapi.Authenticated(keys, func(w http.ResponseWriter, r *http.Request) error {
		r.Header.Set("X-User-ID", httpapi.User(r).Subject)
		native.Routes().ServeHTTP(w, r)
		return nil
	}))))
	return mux, db.Close, nil
}
