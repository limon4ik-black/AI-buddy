package httpapi

import (
	"context"
	"net/http"
	"time"
)

func Health(mux *http.ServeMux, db interface{ Ping(context.Context) error }, name string) {
	mux.HandleFunc("GET /healthz", Handle(func(w http.ResponseWriter, r *http.Request) error {
		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()
			if err := db.Ping(ctx); err != nil {
				return err
			}
		}
		JSON(w, 200, map[string]string{"status": "ok", "service": name})
		return nil
	}))
}
