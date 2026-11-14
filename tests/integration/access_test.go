package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	info "github.com/limon4ik-black/AI-buddy/employee-service/employee"
	"github.com/limon4ik-black/AI-buddy/shared/contracts"
	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
)

func TestAccessRequest(t *testing.T) {
	h := setup(t)
	key := "idempotent-access-key-0001"
	send := func(system string, confirmed bool, key string, status int) []byte {
		return req(t, h.client, h.gateway, "POST", "/api/v1/access-requests", h.token, obj("system", system, "confirmed", confirmed), status, "Idempotency-Key", key)
	}
	send("stash", false, key, 400)
	send("stash", true, "", 400)
	send("secret-service", true, key, 403)
	send("stash", true, key, 201)
	r := unpack[info.AccessRequest](t, send("stash", true, key, 200))
	if r.ExternalID == "" || r.Status != "submitted" {
		t.Fatal(r)
	}
	send("vpn", true, key, 409)
	req(t, h.client, h.gateway, "GET", "/api/v1/access-requests", h.token, nil, 200)
	list := unpack[contracts.Checklist](t, req(t, h.client, h.gateway, "GET", "/api/v1/checklist", h.token, nil, 200))
	if list.Completed != 0 {
		t.Fatal("Submitting request completed checklist")
	}
	// Stable provider result: retries (including restarts) create no second side effect.
	a := req(t, h.client, h.corpURL, "POST", "/internal/v1/access-requests", "", contracts.CorpInput{UserID: "x", System: "stash", IdempotencyKey: key}, 201, "X-Service-Token", "corp-secret")
	b := req(t, h.client, h.corpURL, "POST", "/internal/v1/access-requests", "", contracts.CorpInput{UserID: "x", System: "stash", IdempotencyKey: key}, 201, "X-Service-Token", "corp-secret")
	if !bytes.Equal(a, b) {
		t.Fatal("Mock provider is not idempotent")
	}
	req(t, h.client, h.corpURL, "POST", "/internal/v1/access-requests", "", contracts.CorpInput{}, 400, "X-Service-Token", "corp-secret")
	h.info.CorpURL = "http://127.0.0.1:1"
	send("vpn", true, key+"new", 502)
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpapi.JSON(w, 200, contracts.CorpResult{ID: "", Status: "granted"})
	}))
	defer bad.Close()
	h.info.CorpURL = bad.URL
	send("vpn", true, key+"new", 502)
}
